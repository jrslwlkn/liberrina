package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"liberrina/db/generated"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html/charset"
)

var db *sql.DB
var query *queries.Queries
var ctx context.Context
var templs = make(map[string]*template.Template)
var fs http.Handler

func main() {
	var err error
	db, err = sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	query = queries.New(db)
	ctx = context.Background()

	templs["404"] = template.Must(template.ParseFiles("www/page-layout.html", "www/404.html"))
	templs["index"] = template.Must(template.ParseFiles("www/page-layout.html", "www/index.html"))
	templs["add-lang"] = template.Must(template.ParseFiles("www/page-layout.html", "www/add-lang.html"))
	templs["add-doc"] = template.Must(template.ParseFiles("www/page-layout.html", "www/add-doc.html"))
	templs["db-error"] = template.Must(template.ParseFiles("www/db-error.html"))
	templs["app-error"] = template.Must(template.ParseFiles("www/app-error.html"))
	templs["success"] = template.Must(template.ParseFiles("www/success.html"))
	templs["doc"] = template.Must(template.ParseFiles("www/page-layout.html", "www/doc.html"))

	fs = http.StripPrefix("/www/", http.FileServer(http.Dir("www")))
	http.HandleFunc("/www/", handleStatic)

	http.HandleFunc("/add-lang", handleAddLang)
	http.HandleFunc("/add-doc", handleAddDoc)
	http.HandleFunc("/doc/", handleDoc)
	http.HandleFunc("/", handleIndex)

	log.Println("listening...")
	log.Fatal(http.ListenAndServe(":6969", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		render(w, "404", nil)
		return
	}
	docs, err := query.GetDocs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	langs, err := query.GetLangs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	render(w, "index", IndexData{docs, langs})
}

func handleDoc(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		render(w, "404", nil)
		return
	}
	val, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		render(w, "404", nil)
		return
	}
	doc, err := query.GetDocByID(ctx, val)
	if err != nil {
		render(w, "404", nil)
		return
	}
	render(w, "doc", doc)
}

func handleAddLang(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		langs, err := query.GetAllLangs(ctx)
		if err != nil {
			log.Fatal(err)
		}
		render(w, "add-lang", langs)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		form := r.Form
		lookupURI1 := sql.NullString{String: form.Get("lookup_uri_1"), Valid: true}
		if lookupURI1.String != "" {
			lookupURI1.String = form.Get("lookup_uri_1")
		}
		lookupURI2 := sql.NullString{String: form.Get("lookup_uri_2"), Valid: true}
		if lookupURI2.String != "" {
			lookupURI2.String = form.Get("lookup_uri_2")
		}
		charsPattern := form.Get("chars_pattern")
		if charsPattern == "" {
			charsPattern = `[^A-Za-z'-]`
		}
		sentenceSep := form.Get("sentence_sep")
		if sentenceSep == "" {
			sentenceSep = `[.\?!;]`
		}
		_, err := query.AddLang(ctx, queries.AddLangParams{
			Name:           form.Get("name"),
			FromID:         form.Get("from_id"),
			ToID:           form.Get("to_id"),
			QuickLookupURI: form.Get("quick_lookup_uri"),
			LookupURI1:     lookupURI1,
			LookupURI2:     lookupURI2,
			CharsPattern:   charsPattern,
			SentenceSep:    sentenceSep,
		})
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

func handleAddDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		langs, err := query.GetAllLangs(ctx)
		if err != nil {
			log.Fatal(err)
		}
		render(w, "add-doc", langs)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		form := r.Form
		fmt.Println(form)

		// author := sql.NullString{String: form.Get("author"), Valid: true}
		// tags := sql.NullString{String: form.Get("tags"), Valid: true}
		// notes := sql.NullString{String: form.Get("notes"), Valid: true}

		body := form.Get("doc_body")
		// fmt.Println("body is: ", body)
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
			var chunks []queries.AddChunkParams
			var cur queries.AddChunkParams
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
							chunks = append(chunks, cur)
							builder.Reset()
							i++
						}
						// create another particle
						cur = queries.AddChunkParams{Position: i, DocID: 0} // TODO: use actual document ID
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
			chunks = append(chunks, cur)

			if err := query.PruneChunks(ctx, 0); err != nil { // TODO: use actual document ID
				log.Fatal(err)
			}

			start := time.Now()
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			qtx := query.WithTx(tx)
			for _, chunk := range chunks { // TODO: any better way to avoid storing all these in memory?
				err = qtx.PruneChunks(ctx, 0)
				if err != nil {
					log.Fatal(err)
				}
				_, err := qtx.AddChunk(ctx, chunk)
				if err != nil {
					log.Fatal(err)
				}
				// fmt.Printf("i: %d, value: '%s', suffix: '%s'\n", x.Index, x.Value, x.Suffix)
			}
			if err := tx.Commit(); err != nil {
				log.Fatal(err)
			}
			log.Println("inserted in ", time.Since(start))
		}
	}
}

func render(w http.ResponseWriter, name string, data interface{}) {
	t, ok := templs[name]
	if !ok || name == "404" {
		w.WriteHeader(http.StatusNotFound)
		templs["404"].ExecuteTemplate(w, "404", nil)
		return
	}
	t.ExecuteTemplate(w, name, data)
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".ico") || strings.HasSuffix(r.URL.Path, ".jpg") {
		fs.ServeHTTP(w, r)
	} else {
		render(w, "404", nil)
	}
}
