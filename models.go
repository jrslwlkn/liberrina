package main

import "time"

type LangDim struct {
	Id string `db:"id"`
	Name string `db:"name"`
}

type Document struct {
	Id int64 `db:"doc_id"`
	Title string `db:"title"`
	Author string `db:"author"`
	Body string `db:"body"`
	AddedAt time.Time `db:"added_at"`
	TotalTermCount int64 `db:"term_count"`
	NewTermCount int64 `db:"terms_new"`
	SentenceCount int64 `db:"sentence_count"`
}

