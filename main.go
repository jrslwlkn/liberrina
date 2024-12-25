package main

import (
	"context"
	"database/sql"
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
	docs, err := query.GetDocs(ctx, 0) // TODO: add users
	if err != nil {
		render(w, "db-error", err.Error())
		log.Println("db error when trying to get docs in index:", err.Error())
		return
	}
	for i := range docs {
		if docs[i].Author == "" {
			docs[i].Author = "[No Author]"
		}
	}
	langs, err := query.GetLangs(ctx, 0) // TODO: add users
	if err != nil {
		render(w, "db-error", err.Error())
		log.Println("db error when trying to get langs in index:", err.Error())
		return
	}
	log.Println("rendering index")
	render(w, "index", IndexData{docs, langs})
}

func handleDoc(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	log.Println("requested doc path:", r.URL.Path)
	if len(parts) != 3 {
		render(w, "404", nil)
		return
	}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		render(w, "404", nil)
		return
	}
	doc, err := query.GetDoc(ctx, id)
	if err != nil {
		render(w, "404", nil)
		return
	}
	log.Println("rendering doc", doc.DocID)
	render(w, "doc", doc)
}

func handleAddLang(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		langs, err := query.GetAllLangs(ctx)
		if err != nil {
			log.Println("db error when trying to get all langs (adding new):", err.Error())
			render(w, "db-error", err.Error())
		}
		render(w, "add-lang", langs)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		form := r.Form
		charsPattern := form.Get("chars_pattern")
		if charsPattern == "" {
			charsPattern = `[^A-Za-z'-]`
		}
		sentenceSep := form.Get("sentence_sep")
		if sentenceSep == "" {
			sentenceSep = `[.\?!;]`
		}
		addedID, err := query.AddLang(ctx, queries.AddLangParams{
			Name:           form.Get("name"),
			FromID:         form.Get("from_id"),
			ToID:           form.Get("to_id"),
			QuickLookupURI: form.Get("quick_lookup_uri"),
			LookupURI1:     form.Get("lookup_uri_1"),
			LookupURI2:     form.Get("lookup_uri_2"),
			CharsPattern:   charsPattern,
			SentenceSep:    sentenceSep,
			UserID:         0, // TODO
		})
		if err != nil {
			log.Println("db error when adding lang:", err.Error())
			render(w, "db-error", err.Error())
		} else {
			log.Println("added lang", addedID)
			render(w, "success", nil)
		}
	}
}

func handleAddDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		langs, err := query.GetLangs(ctx, 0)
		if err != nil {
			log.Println("db error when trying to get langs (adding doc):", err.Error())
			render(w, "db-error", err.Error())
		}
		render(w, "add-doc", langs)
	} else if r.Method == http.MethodPost {
		r.ParseMultipartForm(20 << 20) // 20 MB limit
		form := r.Form

		body := strings.TrimSpace(form.Get("doc_body"))
		if body == "" {
			file, _, err := r.FormFile("doc_file")
			if err != nil {
				log.Println("error when getting form field doc_file:", err.Error())
				render(w, "app-error", err.Error())
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
			file.Seek(0, 0)

			reader, err := charset.NewReaderLabel(enc, file)
			if err != nil {
				log.Println("error when creating new reader label for charset (adding doc):", err.Error())
				render(w, "app-error", err.Error())
				return
			}
			fileBytes, err := io.ReadAll(reader)
			if err != nil {
				log.Println("error when reading full body (adding doc):", err.Error())
				render(w, "app-error", err.Error())
				return
			}

			body = strings.TrimSpace(string(fileBytes))
		}
		re := regexp.MustCompile(`[A-Za-z'А-Яа-я'ґЃєЄїЇіІ]`)

		var i int64 = 1
		var cur queries.AddChunkParams
		var inTerm bool = false
		var builder strings.Builder

		start := time.Now()
		tx, err := db.Begin()
		if err != nil {
			log.Println("db error when trying to start a transaction:", err.Error())
			render(w, "db-error", err.Error())
			return
		}
		defer tx.Rollback()

		qtx := query.WithTx(tx)

		langID, err := strconv.Atoi(form.Get("lang_id"))
		if err != nil {
			log.Println("error when trying to get int value lang_id from from (adding doc):", err.Error())
			render(w, "app-error", err.Error())
			return
		}

		docID, err := qtx.AddDoc(ctx, queries.AddDocParams{
			Title:  form.Get("title"),
			Author: form.Get("author"),
			Body:   body,
			Notes:  form.Get("notes"),
			LangID: int64(langID),
			UserID: 0, // TODO: add users
		})
		if err != nil {
			log.Println("db error when trying to add doc:", err.Error())
			render(w, "db-error", err.Error())
			return
		}

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
						err := qtx.AddChunk(ctx, cur)
						if err != nil {
							log.Println("db error when trying to add chunk:", err.Error())
							render(w, "db-error", err.Error())
							return
						}
						builder.Reset()
						i++
					}
					// create another particle
					cur = queries.AddChunkParams{Position: i, DocID: docID}
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
		err = qtx.AddChunk(ctx, cur)
		if err != nil {
			log.Println("db error when trying to add chunk at the very end:", err.Error())
			render(w, "db-error", err.Error())
			return
		}

		// FIXME: this is for testing only
		// if err := query.PruneChunks(ctx, 0); err != nil {
		// 	log.Fatal(err)
		// }

		err = qtx.AddTerms(ctx, docID)
		if err != nil {
			log.Println("db error when adding terms (adding doc):", err.Error())
			render(w, "db-error", err.Error())
			return
		}

		if err := tx.Commit(); err != nil {
			log.Println("db error when trying to commit transaction for adding doc:", err.Error())
			render(w, "db-error", err.Error())
			return
		}

		log.Println("inserted", i, "in", time.Since(start))
		render(w, "success", nil)
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
	if strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".ico") || strings.HasSuffix(r.URL.Path, ".jpg") || strings.HasSuffix(r.URL.Path, ".js") {
		fs.ServeHTTP(w, r)
	} else {
		render(w, "404", nil)
	}
}
