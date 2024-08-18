package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html/charset"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	fmt.Println("listening...")

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add-lang", handleAddLang)
	http.HandleFunc("/add-doc", handleAddDoc)
	http.HandleFunc("/doc/", handleDoc)

	log.Println("listening...")
	log.Fatal(http.ListenAndServe(":6969", nil))
}

type IndexData struct {
	Docs []Document
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	var docs []Document
	_ = query("select doc_id, title, author, added_at, term_count, terms_new, sentence_count from docs",
		nil,
		docs,
	)
	// var langs []Lang
	// _ = sql2slice("select lang_id, name, ",
	// 	&docs,
	// )
	// for i, d := range docs {
	// 	fmt.Printf("%d - %s - %s - %s - %d \n", i, d.Title, d.Author, d.AddedAt, d.NewTermCount)
	// }
	temp := template.Must(template.ParseFiles("www/index.html", "www/base.html"))
	temp.ExecuteTemplate(w, "base", IndexData{Docs: docs})
}

func handleDoc(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		template.Must(template.ParseFiles("www/404.html")).Execute(w, nil)
		return
	}
	val, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		template.Must(template.ParseFiles("www/404.html")).Execute(w, nil)
		return
	}
	var docs []Document
	_ = query("select doc_id, title, author, added_at, term_count, terms_new, sentence_count from docs where doc_id = ?",
		[]any{val},
		docs,
	)
	if len(docs) != 1 {
		template.Must(template.ParseFiles("www/404.html")).Execute(w, nil)
		return
	}
	temp := template.Must(template.ParseFiles("www/doc.html", "www/base.html"))
	temp.ExecuteTemplate(w, "base", docs[0])
}

func handleAddLang(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var langs []LangDim
		err := query("select id, name from langs_dim", nil, langs)
		if err != nil {
			log.Fatal(err)
			return
		}
		temp := template.Must(template.ParseFiles("www/add-lang.html", "www/form-styles.html", "www/base.html"))
		temp.ExecuteTemplate(w, "base", langs)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		form := r.Form
		lookupURI2 := sql.NullString{String: form.Get("lookup_uri_2"), Valid: true}
		if lookupURI2.String != "" {
			lookupURI2.String = form.Get("lookup_uri_2")
		}
		charsPattern := sql.NullString{String: form.Get("chars_pattern"), Valid: true}
		if charsPattern.String == "" {
			charsPattern.String = `[^A-Za-z'-]`
		}
		sentenceSep := sql.NullString{String: form.Get("sentence_sep"), Valid: true}
		if sentenceSep.String == "" {
			sentenceSep.String = `[.\?!;]`
		}
		_, err := db.Exec(`insert into langs(
			name,
			from_id,
			to_id,
			quick_lookup_uri,
			lookup_uri_1,
			lookup_uri_2,
			chars_pattern,
			sentence_sep,
			added_at
		) values (
			?, ?, ?, ?, ?, ?, ?, ?, ?, datetime()
		)`,
			form.Get("name"),
			form.Get("from_id"),
			form.Get("to_id"),
			form.Get("quick_lookup_uri"),
			form.Get("lookup_uri_1"),
			lookupURI2,
			charsPattern,
			sentenceSep,
		)
		if err != nil {
			w.Write([]byte(
				"<div id='result' class='field error'><b>Database Error</b><br><br><code>" +
					err.Error() +
					"</code></div>"))
			fmt.Println("error: ", err.Error())
		} else {
			w.Write([]byte(`
				<div id="result">
					<b>✅ Success!</b></br><br>
					Go <a href="/">home</a>.
				</div>
				<style> .field, button { display: none } </style>`,
			))
		}
	}
}

// Particle represents a small chunk of text like a word and surrounding spaces and punctuation.
type Particle struct {
	Index  int64
	Value  string
	Suffix string
	Level  uint8
}

func handleAddDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var langs []Lang
		err := query("select lang_id, name from langs", nil, langs)
		if err != nil {
			log.Fatal(err)
			return
		}
		temp := template.Must(template.ParseFiles("www/add-doc.html", "www/form-styles.html", "www/base.html"))
		temp.ExecuteTemplate(w, "base", langs)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		form := r.Form
		fmt.Println(form)

		// author := sql.NullString{String: form.Get("author"), Valid: true}
		// tags := sql.NullString{String: form.Get("tags"), Valid: true}
		// notes := sql.NullString{String: form.Get("notes"), Valid: true}

		body := form.Get("doc_body")
		fmt.Println("body is: ", body)
		if body == "" {
			r.ParseMultipartForm(20 << 20) // 20 MB limit
			file, _, err := r.FormFile("doc_file")
			if err != nil {
				log.Fatal(err)
				return
			}
			defer file.Close()

			preview := make([]byte, 1024)
			io.ReadFull(file, preview)
			_, enc, certain := charset.DetermineEncoding(preview, "plain/text")
			// HACK: this is to make Ukrainian work
			if !certain && enc == "windows-1252" {
				enc = "windows-1251"
			}

			reader, err := charset.NewReaderLabel(enc, file)
			if err != nil {
				log.Fatal(err)
			}
			fileBytes, err := io.ReadAll(reader)
			if err != nil {
				log.Fatal(err)
			}

			body := string(fileBytes)
			re := regexp.MustCompile(`[A-Za-z'А-Яа-я'ґЃєЄїЇіІ]`)

			var i int64 = 1
			var particles []Particle
			var cur Particle
			var inTerm bool = false
			var builder strings.Builder

			for _, x := range strings.Split(body, "") {
				if re.MatchString(x) {
					if inTerm {
						// NOTE: keep growing the term
						builder.WriteString(x)
					} else {
						inTerm = true
						// NOTE: this is the end of suffix
						cur.Suffix = builder.String()
						if cur.Value != "" || cur.Suffix != "" {
							particles = append(particles, cur)
							builder.Reset()
							i++
						}
						// create another particle
						cur = Particle{Index: i}
						builder.WriteString(x)
					}
				} else {
					if inTerm {
						// NOTE: this is the end of the term
						inTerm = false
						cur.Value = builder.String()
						builder.Reset()
						builder.WriteString(x)
					} else {
						// NOTE: keep growing the suffix
						builder.WriteString(x)
					}
				}
			}

			if inTerm {
				cur.Value = builder.String()
			} else {
				cur.Suffix = builder.String()
			}
			particles = append(particles, cur)

			_, err = db.Exec("delete from chunks where doc_id=0")
			if err != nil {
				fmt.Println(err)
				return
			}

			start := time.Now()
			db.Exec("begin transaction")
			for _, x := range particles {
				_, err := db.Exec(
					`insert into chunks (
						doc_id,
						position,
						value,
						suffix
					 ) values (?, ?, ?, ?)`,
					0,
					x.Index,
					x.Value,
					x.Suffix,
				)
				if err != nil {
					fmt.Println(err)
					return
				}
				// fmt.Printf("i: %d, value: '%s', suffix: '%s'\n", x.Index, x.Value, x.Suffix)
			}
			db.Exec("commit transaction")
			fmt.Println("inserted in ", time.Since(start))
		}
	}
}

func query[T any](query string, args []any, dest []T) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	sliceVal := reflect.ValueOf(dest)
	elemType := sliceVal.Type().Elem()
	for rows.Next() {
		newElem := reflect.New(elemType)
		fields := make([]interface{}, len(columns))
		for i, col := range columns {
			field, found := elemType.FieldByNameFunc(func(fieldName string) bool {
				field, _ := elemType.FieldByName(fieldName)
				return field.Tag.Get("db") == col
			})
			if found {
				fields[i] = newElem.FieldByIndex(field.Index).Addr().Interface()
			} else {
				var placeholder interface{}
				fields[i] = &placeholder
			}
		}

		if err := rows.Scan(fields...); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}

	return nil
}
