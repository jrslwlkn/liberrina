package main

import "time"
import queries "liberrina/db/generated"

type LangDim struct {
	Id   string `db:"id"`
	Name string `db:"name"`
}

type Lang struct {
	Id   string `db:"lang_id"`
	Name string `db:"name"`
}

type Doc struct {
	Id             int64     `db:"doc_id"`
	Title          string    `db:"title"`
	Author         string    `db:"author"`
	Body           string    `db:"body"`
	AddedAt        time.Time `db:"added_at"`
	TotalTermCount int64     `db:"term_count"`
	NewTermCount   int64     `db:"terms_new"`
	SentenceCount  int64     `db:"sentence_count"`
}

type IndexData struct {
	Docs  []queries.GetDocsRow
	Langs []queries.GetLangsRow
}

// Particle represents a small chunk of text like a word and surrounding spaces and punctuation.
type Particle struct {
	Index  int64
	Value  string
	Suffix string
	Level  uint8
}
