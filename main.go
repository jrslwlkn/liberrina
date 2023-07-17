package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("INSERT INTO langs_dim(id, name) VALUES(?, ?)")
	defer stmt.Close()

	_, err = stmt.Exec("ua", "Ukrainian")
	_, err = stmt.Exec("la", "Latin")

	tx.Commit()

	rows, _ := db.Query("SELECT * FROM langs_dim")
	defer rows.Close()
	for rows.Next() {
		var id string
		var name string
		_ = rows.Scan(&id, &name)
		fmt.Println(id, "-", name)
	}

	fmt.Println("hello world")
	
	h1 := func(w http.ResponseWriter, r *http.Request){
		temp := template.Must(template.ParseFiles("html/index.html"))
		temp.Execute(w, nil)
	}
	http.HandleFunc("/", h1)
	
	log.Fatal(http.ListenAndServe(":6942", nil))
}
