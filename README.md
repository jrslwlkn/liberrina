# Liberrina

(WIP) Simple web app for learning languages/words by reading with built-in translation capabilities via Google Translate and Anki integration via AnkiConnect.

## Development Setup

1. Install [golang](https://go.dev/dl/).

1. Install [sqlite3](https://www.sqlite.org/download.html).

1. Install [goose](https://github.com/pressly/goose).

1. Install [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html).

1. Run migrations:
    ```
    cd db
    goose -dir migrations sqlite3 app.db up
    ```
1. Generate Go from SQL queries (if queries are added/updated in `db/queries.sql`)

    ```
    cd db
    sqlc generate
    ```

1. Start dev server:
    ```
    go run *.go
    ```

    Or use hot reloading with [entr](https://github.com/eradman/entr):
	```
    find . -name "*.go" | entr -r go run . 
    ```

    Or use hot reloading with [air](https://github.com/air-verse/air):

    ```
    air
    ```
