package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/datasektionen/GOrdian/internal/config"
	"github.com/datasektionen/GOrdian/internal/database"

	"github.com/datasektionen/GOrdian/internal/web"
)

func main() {
	envVar := config.GetEnv()
	dbGO, err := database.Connect(envVar.PsqlconnStringGOrdian)
	if err != nil {
		log.Printf("error accessing GOrdian database: %v", err)
	}

	dbCF, err := database.Connect(envVar.PsqlconnStringCashflow)
	if err != nil {
		log.Printf("error accessing Cashflow database: %v", err)
	}

	if err := web.Mount(http.DefaultServeMux, web.Databases{DBCF: dbCF, DBGO: dbGO}); err != nil {
		panic(err)
	}
	panic(http.ListenAndServe("0.0.0.0:3000", nil))
}
