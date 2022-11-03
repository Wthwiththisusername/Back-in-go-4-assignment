package main

import (
	"com.aitu.snippetbox/internal/models"
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	db, dberr := pgxpool.Connect(context.Background(), "postgres://postgres:12345@localhost:5432/snippetbox")
	if dberr != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", dberr)
		os.Exit(1)
	}
	defer db.Close()
	var greeting string
	dberr := db.QueryRow(context.Background(), "DataBase connected").Scan(&greeting)
	if dberr != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", dberr)
		os.Exit(1)
	}
	fmt.Println(greeting)
	// new command line with name 'addr' & default value 4000
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	// connection pool closed
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: templateCache,
	}
	// new http.Server struct
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
