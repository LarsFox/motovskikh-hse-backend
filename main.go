package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "cw_user:cw_password@tcp(127.0.0.1:3306)/coursework?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("connected to mysql")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.Handle("/", http.FileServer(http.Dir("internal/web")))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
