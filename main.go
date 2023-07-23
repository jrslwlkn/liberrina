package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"log"
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


	rows, _ := db.Query("SELECT * FROM langs_dim")
	defer rows.Close()
	for rows.Next() {
		var id string
		var name string
		_ = rows.Scan(&id, &name)
		fmt.Println(id, "-", name)
	}

	log.Fatal(http.ListenAndServe(":6969", nil))
}
	
func handleIndex(w http.ResponseWriter, r *http.Request) {
		temp := template.Must(template.ParseFiles("html/index.html"))
		temp.Execute(w, nil)
	}
}
