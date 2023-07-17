package main

import (
	"fmt"
	"html/template"
	"net/http"
	"log"
)

func main() {
	fmt.Println("hello world")
	
	h1 := func(w http.ResponseWriter, r *http.Request){
		temp := template.Must(template.ParseFiles("html/index.html"))
		temp.Execute(w, nil)
	}
	http.HandleFunc("/", h1)
	
	log.Fatal(http.ListenAndServe(":6942", nil))
}
