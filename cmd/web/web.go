package main

import (
	"github.com/datasektionen/GOrdian/internal/database"
	"log"
	"net/http"

	"github.com/datasektionen/GOrdian/internal/web"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Printf("error accessing database: %v", err)
	}
	if err := web.Mount(http.DefaultServeMux, db); err != nil {
		panic(err)
	}
	panic(http.ListenAndServe("0.0.0.0:3000", nil))
}
