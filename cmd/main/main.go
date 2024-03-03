package main

import (
	"fmt"
	"github.com/datasektionen/GOrdian/internal/parse"
	"log"
)

func main() {
	fmt.Println("You have very many money")
	budget := "test/Budget_2024.xlsx"
	err := parse.ReadExcel(budget)
	if err != nil {
		log.Printf("error parsing Excel file: %v", err)
	}
}
