package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ReportLine struct {
	ReportLineCostCentre          string
	ReportLineSecondaryCostCentre string
	ReportLineBudgetLine          string
	ReportLineTotal               string
}

type ReportBudgetLine struct {
	BudgetLineName string
	Total          string
}

type ReportSecondaryCostCentreLine struct {
	SecondaryCostCentreName string
	BudgetLinesList         []ReportBudgetLine
	Total                   string
}

type ReportCostCentreLine struct {
	CostCentreName           string
	SecondaryCostCentresList []ReportSecondaryCostCentreLine
	Total                    string
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

	CCList, err := getCCList(databases.DBCF)
	if err != nil {
		return fmt.Errorf("failed get scan CCList information from database: %v", err)
	}

	reportLines, err := getReportLines(databases.DBCF, selectedYear, r.FormValue("cc"))
	if err != nil {
		return fmt.Errorf("failed to get scan report lines information from database: %v", err)
	}

	structuredReport, err := StructureReportLines(reportLines)
	if err != nil {
		return fmt.Errorf("failed to structure report lines: %v", err)
	}
	years := getYearsSince2017()

	if err := templates.ExecuteTemplate(w, "report.gohtml", map[string]any{
		"motd":         motdGenerator(),
		"reportLines":  reportLines,
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

func getReportLines(db *sql.DB, year string, cc string) ([]ReportLine, error) {

	var result *sql.Rows
	var err error

	result, err = db.Query(CombinedReportLinesGetStatementStatic, year, cc)
	if err != nil {
		return nil, fmt.Errorf("failed to get reportlines from database: %v", err)
	}
	defer result.Close()

	var reportLines []ReportLine

	for result.Next() {
		var reportLine ReportLine

		err := result.Scan(
			&reportLine.ReportLineCostCentre,
			&reportLine.ReportLineSecondaryCostCentre,
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

// Add cost centre or return existing
func findOrAddCostCentre(costCentres *[]ReportCostCentreLine, costCentreName string) *ReportCostCentreLine {
	for i := range *costCentres {
		if (*costCentres)[i].CostCentreName == costCentreName {
			return &(*costCentres)[i]
		}
	}
	// Add new cost centre if not found
	newCostCentre := ReportCostCentreLine{
		CostCentreName:           costCentreName,
		SecondaryCostCentresList: []ReportSecondaryCostCentreLine{},
		Total:                    "0 kr",
	}
	*costCentres = append(*costCentres, newCostCentre)
	return &(*costCentres)[len(*costCentres)-1]
}

// Add secondary cost centre or return existing
func findOrAddSecondaryCostCentre(secCostCentres *[]ReportSecondaryCostCentreLine, secCostCentreName string) *ReportSecondaryCostCentreLine {
	for i := range *secCostCentres {
		if (*secCostCentres)[i].SecondaryCostCentreName == secCostCentreName {
			return &(*secCostCentres)[i]
		}
	}
	// Add new secondary cost centre if not found
	newSecCostCentre := ReportSecondaryCostCentreLine{
		SecondaryCostCentreName: secCostCentreName,
		BudgetLinesList:         []ReportBudgetLine{},
		Total:                   "0 kr",
	}
	*secCostCentres = append(*secCostCentres, newSecCostCentre)
	return &(*secCostCentres)[len(*secCostCentres)-1]
}

// Function to organize ReportLines into structured data
func StructureReportLines(reportLines []ReportLine) ([]ReportCostCentreLine, error) {
	var costCentres []ReportCostCentreLine

	// Loop through each ReportLine and structure it
	for _, line := range reportLines {
		// Find or add the CostCentre
		costCentre := findOrAddCostCentre(&costCentres, line.ReportLineCostCentre)

		// Find or add the SecondaryCostCentre within the CostCentre
		secCostCentre := findOrAddSecondaryCostCentre(&costCentre.SecondaryCostCentresList, line.ReportLineSecondaryCostCentre)

		// Add the BudgetLine to the SecondaryCostCentre
		secCostCentre.BudgetLinesList = append(secCostCentre.BudgetLinesList, ReportBudgetLine{
			BudgetLineName: line.ReportLineBudgetLine,
			Total:          line.ReportLineTotal,
		})
		var err1, err2 error

		// Update totals for SecondaryCostCentre and CostCentre
		secCostCentre.Total, err1 = addTotals(secCostCentre.Total, line.ReportLineTotal)
		costCentre.Total, err2 = addTotals(costCentre.Total, line.ReportLineTotal)

		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("failed to update totals for SCC or CC: %v%v", err1, err2)
		}
	}

	return costCentres, nil
}

// Helper function to add totals (handling currency and decimal values)
func addTotals(total1, total2 string) (string, error) {
	t1, err1 := strconv.ParseFloat(strings.ReplaceAll(total1, " kr", ""), 64)
	t2, err2 := strconv.ParseFloat(strings.ReplaceAll(total2, " kr", ""), 64)

	if err1 != nil || err2 != nil {
		return "0 kr", fmt.Errorf("failed operate on total float: %v%v", err1, err2)
	}

	return fmt.Sprintf("%.2f kr", t1+t2), nil
}
