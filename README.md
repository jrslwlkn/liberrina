### Development Setup

1. Install [golang](https://go.dev/dl/).

1. Install [sqlite3](https://www.sqlite.org/download.html).

1. Install [goose](https://github.com/pressly/goose).

1. Run migrations:
    ```
    goose -dir migrations sqlite3 app.db up
    ```

1. Start dev server:
    ```
    go run main.go
    ```

	
