package database

import (
	"database/sql"
	"fmt"
	"github.com/datasektionen/GOrdian/internal/excel"
	_ "github.com/lib/pq"
)

// TODO sanitize integer values in parser
func InsertBudget(db *sql.DB, budgetLines []excel.BudgetLine) error {
	var insertStatementDyn = `insert into "budget_lines"("id", "name", "income", "expenses", "comment", "account", "secondary_cost_centre_id") values($1, $2, $3, $4, $5, $6, $7)`
	fmt.Println(insertStatementDyn)
	_, err := db.Exec(insertStatementDyn, 1, budgetLines[0].BudgetLineName, 6, 9, budgetLines[0].BudgetLineComment, budgetLines[0].BudgetLineAccount, 1)
	if err != nil {
		return fmt.Errorf("failed to insert to database: %v", err)
	}
	return nil
}

func WipeDatabase(db *sql.DB) error {
	var truncateStatement = `TRUNCATE "budget_lines" , "cost_centres", "secondary_cost_centres"`
	_, err := db.Exec(truncateStatement)
	if err != nil {
		return fmt.Errorf("failed to truncate database: %v", err)
	}
	return nil
}
