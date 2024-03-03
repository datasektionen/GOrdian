package main

import (
	"fmt"
	"github.com/datasektionen/GOrdian/internal/parse"
)

func main() {
	fmt.Println("You have very many money")
	budget := "test/Budget_2024.xlsx"
	err := parse.ReadExcel(budget)
	if err != nil {
		_ = fmt.Errorf("failed to parse Excel file: %v", err)
	}
}
