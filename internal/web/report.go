package web

import (
	"database/sql"
	"fmt"
	"net/http"
)

type ReportLine struct {
	ReportLineCC         string
	ReportLineSCC        string
	ReportLineBudgetLine string
	ReportLineTotal      string
}

func reportPage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {

	reportLines, err := getReportLines(db)
	if err != nil {
		return fmt.Errorf("failed get scan report lines information from database: %v", err)
	}

	if err := templates.ExecuteTemplate(w, "report.gohtml", map[string]any{
		"motd":        motdGenerator(),
		"reportLines": reportLines,
		"permissions": perms,
		"loggedIn":    loggedIn,
	}); err != nil {
		return fmt.Errorf("Could not render template: %w", err)
	}
	return nil
}

func getReportLines(db *sql.DB) ([]ReportLine, error) {
	var reportLinesGetStatementStatic = `
		SELECT 
		    cost_centre, 
		    secondary_cost_centre, 
		    budget_line, 
		    SUM(amount) AS total_amount
		FROM (
		SELECT 
		    ep.cost_centre, 
		    ep.secondary_cost_centre, 
		    ep.budget_line, 
		    ep.amount, 
		    EXTRACT(YEAR FROM e.expense_date) AS date, 
		    e.description
		FROM expenses_expensepart AS ep
		INNER JOIN expenses_expense AS e ON ep.expense_id = e.id
		UNION ALL
		SELECT 
		    ip.cost_centre, 
		    ip.secondary_cost_centre, 
		    ip.budget_line, 
		    ip.amount, 
		    EXTRACT(YEAR FROM (COALESCE(i.invoice_date, i.payed_at))) AS invoice_date, 
		    i.description
		FROM invoices_invoicepart AS ip
		INNER JOIN invoices_invoice AS i ON ip.invoice_id = i.id
		) AS combined
		WHERE COALESCE(date = 2024, true)
		GROUP BY cost_centre, secondary_cost_centre, budget_line
		ORDER BY cost_centre, secondary_cost_centre, budget_line;
	`
	result, err := db.Query(reportLinesGetStatementStatic)
	if err != nil {
		return nil, fmt.Errorf("failed to get reportlines from database: %v", err)
	}

	var reportLines []ReportLine
	for result.Next() {
		var reportLine ReportLine

		err := result.Scan(
			&reportLine.ReportLineCC,
			&reportLine.ReportLineSCC,
			&reportLine.ReportLineBudgetLine,
			&reportLine.ReportLineTotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report line from query result: %v", err)
		}
		reportLines = append(reportLines, reportLine)
	}
	return reportLines, nil
}
