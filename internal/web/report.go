package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SimpleBudgetLine struct {
	BudgetLineCostCentreName          string
	BudgetLineSecondaryCostCentreName string
	BudgetLineName         			  string
	BudgetLineExpense             	  string
}

type CashflowLine struct {
	CashflowLineCostCentre          string
	CashflowLineSecondaryCostCentre string
	CashflowLineBudgetLine          string
	CashflowLineTotal               string
}

type ReportBudgetLine struct {
	BudgetLineName 	string
	Total          	string
	Budget			string
}

type ReportSecondaryCostCentreLine struct {
	SecondaryCostCentreName string
	BudgetLinesList         []ReportBudgetLine
	Total                   string
	Budget					string
}

type ReportCostCentreLine struct {
	CostCentreName           string
	SecondaryCostCentresList []ReportSecondaryCostCentreLine
	Total                    string
	Budget					 string
}

func getYearsSince2017() []string {
	startYear := 2017
	currentYear := time.Now().Year()
	var years []string

	for year := startYear; year <= currentYear; year++ {
		years = append(years, strconv.Itoa(year))
	}

	return years
}

func reportPage(w http.ResponseWriter, r *http.Request, databases Databases, perms []string, loggedIn bool) error {

	currentYear := strconv.Itoa(time.Now().Year())

	selectedYear := r.FormValue("year")
	if selectedYear == "" {
		selectedYear = currentYear
	}

	// Fetch simple budget lines
	simpleBudgetLines, err := getSimpleBudgetLines(databases.DBGO)
	if err != nil {
		return fmt.Errorf("failed to get simple budget line information from database: %v", err)
	}

	CCList, err := getCCList(databases.DBCF)
	if err != nil {
		return fmt.Errorf("failed get scan CCList information from database: %v", err)
	}

	// Fetch cashflow lines
	cashflowLines, err := getCashflowLines(databases.DBCF, selectedYear, r.FormValue("cc"))
	if err != nil {
		return fmt.Errorf("failed to get scan cashflow lines information from database: %v", err)
	}

	// Populate report lines with expenses only for the current year
	// reportLines, err = addBudgetToReportLinesForCurrentYear(reportLines, simpleBudgetLines, selectedYear, currentYear)
	// if err != nil {
	// 	return fmt.Errorf("failed to populate report lines for the current year: %v", err)
	// }

	structuredReport, err := StructureReportLines(cashflowLines, simpleBudgetLines, selectedYear)
	if err != nil {
		return fmt.Errorf("failed to structure cashflow and simple budget lines: %v", err)
	}
	years := getYearsSince2017()

	if err := templates.ExecuteTemplate(w, "report.gohtml", map[string]any{
		"motd":         motdGenerator(),
		"cashflowLines":  cashflowLines,
		"permissions":  perms,
		"loggedIn":     loggedIn,
		"report":       structuredReport,
		"CCList":       CCList,
		"years":        years,
		"SelectedCC":   r.FormValue("cc"),
		"SelectedYear": selectedYear,
	}); err != nil {
		return fmt.Errorf("could not render template: %w", err)
	}
	return nil
}

func getCCList(db *sql.DB) ([]string, error) {
	var result *sql.Rows
	var err error

	result, err = db.Query(uniqueCCGetStatementStatic)
	if err != nil {
		return nil, fmt.Errorf("failed to get CCList from database: %v", err)
	}
	defer result.Close()

	var CCList []string

	for result.Next() {
		var CC string

		err := result.Scan(
			&CC,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CC from query result: %v", err)
		}
		CCList = append(CCList, CC)
	}
	return CCList, nil
}

func getCashflowLines(db *sql.DB, year string, cc string) ([]CashflowLine, error) {

	var result *sql.Rows
	var err error

	result, err = db.Query(CombinedCashflowLinesGetStatementStatic, year, cc)
	if err != nil {
		return nil, fmt.Errorf("failed to get cashflow lines from database: %v", err)
	}
	defer result.Close()

	var cashflowLines []CashflowLine

	for result.Next() {
		var cashflowLine CashflowLine

		err := result.Scan(
			&cashflowLine.CashflowLineCostCentre,
			&cashflowLine.CashflowLineSecondaryCostCentre,
			&cashflowLine.CashflowLineBudgetLine,
			&cashflowLine.CashflowLineTotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cashflow line from query result: %v", err)
		}
		cashflowLines = append(cashflowLines, cashflowLine)
	}
	return cashflowLines, nil
}

// Helper function to find or add a CostCentre
func findOrAddCostCentre(costCentres *[]ReportCostCentreLine, name string) *ReportCostCentreLine {
	for i := range *costCentres {
		if (*costCentres)[i].CostCentreName == name {
			return &(*costCentres)[i]
		}
	}
	*costCentres = append(*costCentres, ReportCostCentreLine{
		CostCentreName:           name,
		SecondaryCostCentresList: []ReportSecondaryCostCentreLine{},
		Total:                    "0",
		Budget:                   "0",
	})
	return &(*costCentres)[len(*costCentres)-1]
}

// Helper function to find or add a SecondaryCostCentre
func findOrAddSecondaryCostCentre(secCostCentres *[]ReportSecondaryCostCentreLine, name string) *ReportSecondaryCostCentreLine {
	for i := range *secCostCentres {
		if (*secCostCentres)[i].SecondaryCostCentreName == name {
			return &(*secCostCentres)[i]
		}
	}
	*secCostCentres = append(*secCostCentres, ReportSecondaryCostCentreLine{
		SecondaryCostCentreName: name,
		BudgetLinesList:         []ReportBudgetLine{},
		Total:                   "0",
		Budget:                  "0",
	})
	return &(*secCostCentres)[len(*secCostCentres)-1]
}

// Function to organize CashflowLines into structured data
func StructureReportLines(cashflowLines []CashflowLine, simpleBudgetLines []SimpleBudgetLine, selectedYear string) ([]ReportCostCentreLine, error) {
	var costCentres []ReportCostCentreLine

	currentYear := strconv.Itoa(time.Now().Year())
	//currentYear := "2024"

	// Loop through each CashflowLine and structure it
	for _, line := range cashflowLines {
		// Find or add the CostCentre
		costCentre := findOrAddCostCentre(&costCentres, line.CashflowLineCostCentre)

		// Find or add the SecondaryCostCentre within the CostCentre
		secCostCentre := findOrAddSecondaryCostCentre(&costCentre.SecondaryCostCentresList, line.CashflowLineSecondaryCostCentre)

		// Add the BudgetLine to the SecondaryCostCentre
		secCostCentre.BudgetLinesList = append(secCostCentre.BudgetLinesList, ReportBudgetLine{
			BudgetLineName: line.CashflowLineBudgetLine,
			Total:          line.CashflowLineTotal,
		})
		
		// Update totals for SecondaryCostCentre and CostCentre
		var err1, err2 error

		secCostCentre.Total, err1 = addTotals(secCostCentre.Total, line.CashflowLineTotal)
		costCentre.Total, err2 = addTotals(costCentre.Total, line.CashflowLineTotal)

		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("failed to update totals for SCC or CC: %v%v", err1, err2)
		}
	}

	// Process SimpleBudgetLines
	for _, budgetLine := range simpleBudgetLines {
		// Check if the SimpleBudgetLine exists in the CashflowLines
		existsInCashflow := false
		for _, cashflowLine := range cashflowLines {
			if cashflowLine.CashflowLineCostCentre == budgetLine.BudgetLineCostCentreName &&
				cashflowLine.CashflowLineSecondaryCostCentre == budgetLine.BudgetLineSecondaryCostCentreName &&
				cashflowLine.CashflowLineBudgetLine == budgetLine.BudgetLineName {
				existsInCashflow = true
				break
			}
		}

		if !existsInCashflow {
			continue
		}

		// Find or add the CostCentre
		costCentre := findOrAddCostCentre(&costCentres, budgetLine.BudgetLineCostCentreName)

		// Find or add the SecondaryCostCentre within the CostCentre
		secCostCentre := findOrAddSecondaryCostCentre(&costCentre.SecondaryCostCentresList, budgetLine.BudgetLineSecondaryCostCentreName)

		// Check if the selectedYear matches the current year
		budgetValue := "0"
		if selectedYear == currentYear {
			budgetValue = makePositive(budgetLine.BudgetLineExpense)
		}

		// Add or update the BudgetLine in the SecondaryCostCentre
		found := false
		for i, bl := range secCostCentre.BudgetLinesList {
			if bl.BudgetLineName == budgetLine.BudgetLineName {
				// Update the budget value
				updatedBudget, err := addTotals(secCostCentre.BudgetLinesList[i].Budget, budgetValue)
				if err != nil {
					return nil, fmt.Errorf("failed to update budget value: %v", err)
				}
				secCostCentre.BudgetLinesList[i].Budget = updatedBudget
				found = true
				break
			}
		}
		if !found {
			// Add a new budget line if it doesn't exist
			secCostCentre.BudgetLinesList = append(secCostCentre.BudgetLinesList, ReportBudgetLine{
				BudgetLineName: budgetLine.BudgetLineName,
				Total:          "0", // Placeholder for total
				Budget:         budgetValue,
			})
		}

		// Update the budget totals for the SecondaryCostCentre and CostCentre
		if selectedYear == currentYear {
			var err error
			secCostCentre.Budget, err = addTotals(secCostCentre.Budget, budgetValue)
			if err != nil {
				return nil, fmt.Errorf("failed to update budget total for SCC: %v", err)
			}

			costCentre.Budget, err = addTotals(costCentre.Budget, budgetValue)
			if err != nil {
				return nil, fmt.Errorf("failed to update budget total for CC: %v", err)
			}
		}
	}

	// Set budget values to "-" if they are zero
	for i := range costCentres {
		if costCentres[i].Budget == "0" {
			costCentres[i].Budget = ""
		}
		for j := range costCentres[i].SecondaryCostCentresList {
			if costCentres[i].SecondaryCostCentresList[j].Budget == "0" {
				costCentres[i].SecondaryCostCentresList[j].Budget = ""
			}
		}
	}

	return costCentres, nil
}

func getSimpleBudgetLines(db *sql.DB) ([]SimpleBudgetLine, error) {
	var query = `
		SELECT 
			cost_centres.name,
			secondary_cost_centres.name,
			budget_lines.name,
			budget_lines.expense
		FROM budget_lines
		JOIN secondary_cost_centres 
			ON budget_lines.secondary_cost_centre_id = secondary_cost_centres.id
		JOIN cost_centres 
			ON secondary_cost_centres.cost_centre_id = cost_centres.id
		ORDER BY cost_centres.name, secondary_cost_centres.name, budget_lines.name
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get simple budget lines from database: %v", err)
	}
	defer rows.Close()

	var simpleBudgetLines []SimpleBudgetLine

	// Iterate over the result rows
	for rows.Next() {
		var simpleBudgetLine SimpleBudgetLine

		err := rows.Scan(
			&simpleBudgetLine.BudgetLineCostCentreName,
			&simpleBudgetLine.BudgetLineSecondaryCostCentreName,
			&simpleBudgetLine.BudgetLineName,
			&simpleBudgetLine.BudgetLineExpense,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan simple budget line from query result: %v", err)
		}

		simpleBudgetLines = append(simpleBudgetLines, simpleBudgetLine)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through simple budget lines: %v", err)
	}

	return simpleBudgetLines, nil
}

// Helper function to add totals (no handling of "kr" suffix)
func addTotals(total1, total2 string) (string, error) {
	// Normalize input values
	total1 = strings.TrimSpace(total1)
	total2 = strings.TrimSpace(total2)

	if total1 == "" {
		total1 = "0"
	}
	if total2 == "" {
		total2 = "0"
	}

	// Convert to floats
	t1, err1 := strconv.ParseFloat(total1, 64)
	t2, err2 := strconv.ParseFloat(total2, 64)

	if err1 != nil || err2 != nil {
		return "0", fmt.Errorf("failed to convert totals to float: %v, %v", err1, err2)
	}

	// Add and format the result
	result := t1 + t2
	return formatNumber(result), nil
}

// Helper function to make a value positive (no "kr" suffix)
func makePositive(value string) string {
	value = strings.TrimSpace(value)

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "0"
	}

	if parsed < 0 {
		parsed = -parsed
	}

	return formatNumber(parsed)
}

func formatNumber(value float64) string {
	if value == float64(int(value)) {
		return fmt.Sprintf("%d", int(value)) // No decimals if the value is an integer
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".") // Remove unnecessary zeros
}