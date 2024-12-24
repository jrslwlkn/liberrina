package main

import queries "liberrina/db/generated"

type IndexData struct {
	Docs  []queries.GetDocsRow
	Langs []queries.GetLangsRow
}
