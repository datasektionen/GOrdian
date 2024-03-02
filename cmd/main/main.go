package main

import (
	"fmt"
	"github.com/datasektionen/GOrdian/internal/parse"
)

func main() {
	fmt.Println("You have very many money")
	budget := "test/Budget_2024.xlsx"
	parse.ReadExcel(budget)
}
