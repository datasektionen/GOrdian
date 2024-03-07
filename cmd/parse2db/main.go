package parse2db

import (
	"fmt"
	"github.com/datasektionen/GOrdian/internal/excel"
	"log"
)

func main() {
	fmt.Println("You have very many money")
	budget := "test/Budget_2024.xlsx"
	err := excel.ReadExcel(budget)
	if err != nil {
		log.Printf("error parsing Excel file: %v", err)
	}
}
