package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
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

	log.Fatal(http.ListenAndServe(":6969", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	temp := template.Must(template.ParseFiles("html/index.html", "html/base.html"))
	temp.ExecuteTemplate(w, "base", nil)
}

func handleAddLang(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var langs []LangDim
		err := sql2slice("select id, name from langs_dim", nil, &langs)
		if err != nil {
			log.Fatal(err)
			return
		}
		temp := template.Must(template.ParseFiles("html/add-lang.html", "html/base.html"))
		temp.ExecuteTemplate(w, "base", langs)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		form := r.Form
		lookupURI2 := sql.NullString{String: form.Get("lookup_uri_2"), Valid: true}
		if lookupURI2.String != "" {
			lookupURI2.String = form.Get("lookup_uri_2")
		}
		termSep := sql.NullString{String: form.Get("term_sep"), Valid: true}
		if termSep.String == "" {
			termSep.String = `\s+`
		}
		trimPattern := sql.NullString{String: form.Get("trim_pattern"), Valid: true}
		if trimPattern.String == "" {
			trimPattern.String = `[^A-Za-z'-]`
		}
		sentenceSep := sql.NullString{String: form.Get("sentence_sep"), Valid: true}
		if sentenceSep.String == "" {
			sentenceSep.String = `[.\?!;]`
		}
		userID := sql.NullInt64{Int64: 0, Valid: true}
		if val, err := strconv.ParseInt(form.Get("user_id"), 10, 64); err == nil {
			userID.Int64 = val
		}
		_, err := db.Exec(`insert into langs(
			name,
			from_id,
			to_id,
			quick_lookup_uri,
			lookup_uri_1,
			lookup_uri_2,
			term_sep,
			trim_pattern,
			sentence_sep,
			user_id,
			added_at
		) values (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime()
		)`,
			form.Get("name"),
			form.Get("from_id"),
			form.Get("to_id"),
			form.Get("quick_lookup_uri"),
			form.Get("lookup_uri_1"),
			lookupURI2,
			termSep,
			trimPattern,
			sentenceSep,
			userID,
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

func handleAddDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		temp := template.Must(template.ParseFiles("html/add-doc.html", "html/form-styles.html", "html/base.html"))
		temp.ExecuteTemplate(w, "base", nil)
	} else if r.Method == http.MethodPost {

	}
}

func sql2slice[T any](query string, args []any, dest *[]T) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	sliceVal := reflect.ValueOf(dest).Elem()
	elemType := sliceVal.Type().Elem()
	for rows.Next() {
		newElem := reflect.New(elemType).Elem()
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
