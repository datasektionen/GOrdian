package database

import (
	"database/sql"
	"fmt"
	"github.com/datasektionen/GOrdian/internal/excel"
	_ "github.com/lib/pq"
)

func InsertBudget(db *sql.DB, budgetLines []excel.BudgetLine) error {
	fmt.Println("debug1")
	var insertStatementDyn = `insert into "budget_lines"("id", "name", "income", "expenses", "comment", "account", "secondary_cost_centre_id") values($1, $2, $3, $4, $5, $6, $7)`
	fmt.Println("debug2")
	fmt.Println(insertStatementDyn)
	fmt.Println("debug3")
	_, err := db.Exec(insertStatementDyn, 1, budgetLines[0].BudgetLineName, 6, 9, budgetLines[0].BudgetLineComment, budgetLines[0].BudgetLineAccount, 1)
	if err != nil {
		return fmt.Errorf("failed to insert to database: %v", err)
	}
	fmt.Println("debug4")
	return nil
}
