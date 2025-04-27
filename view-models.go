package main

import queries "liberrina/db/generated"

type IndexData struct {
	Docs  []queries.GetDocsRow
	Langs []queries.GetLangsRow
}

type DocData struct {
	queries.GetDocMetaRow
	Chunks []queries.GetDocBodyRow
}

type TermData struct {
	Term        string `json:"term"`
	Level       int64  `json:"level"`
	Translation string `json:"translation"`
	DocID       int64  `json:"docID"`
}
