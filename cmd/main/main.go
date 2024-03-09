package main

import (
	"fmt"
	"github.com/datasektionen/GOrdian/internal/database"
	"github.com/datasektionen/GOrdian/internal/excel"
	"log"
)

func main() {
	fmt.Println("You have very many money")
	budget := "test/Budget_2024.xlsx"
	budgetLines, err := excel.ReadExcel(budget)
	if err != nil {
		log.Printf("error parsing Excel file: %v", err)
	}

	db, err := database.Connect()
	if err != nil {
		log.Printf("error accessing database: %v", err)
	}

	err = database.WipeDatabase(db)
	if err != nil {
		log.Printf("error inserting budget in database: %v", err)
	}

	err = database.InsertBudget(db, budgetLines)
	if err != nil {
		log.Printf("error inserting budget in database: %v", err)
	}

	err = database.Close(db)
	if err != nil {
		log.Printf("error wiping database: %v", err)
	}

	//fmt.Println(budgetLines)
	//fmt.Println(budgetLines[0].CostCentreName)
}
