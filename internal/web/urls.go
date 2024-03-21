package web

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/datasektionen/GOrdian/internal/excel"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFiles embed.FS

var templates *template.Template

func Mount(mux *http.ServeMux, db *sql.DB) error {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}
	mux.Handle("/static/", http.FileServerFS(staticFiles))
	mux.Handle("/{$}", page(db, index))
	mux.Handle("/costcentre/{costCentreIDPath}", page(db, costCentrePage))

	return nil
}

func page(db *sql.DB, handler func(w http.ResponseWriter, r *http.Request, db *sql.DB) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r, db)
		if err != nil {
			slog.Error("Error from handler", "error", err)
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
		}
	})
}

func index(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	costCentres, err := getCostCentres(db)
	if err != nil {
		return fmt.Errorf("failed get scan cost centre information from database: %v", err)
	}
	committeeCostCentres, projectCostCentres, otherCostCentres, err := splitCostCentresOnType(costCentres)
	if err != nil {
		return fmt.Errorf("failed spit cost centres on type: %v", err)
	}
	if err := templates.ExecuteTemplate(w, "index.html", map[string]any{
		"motd":       "You have very many money",
		"committees": committeeCostCentres,
		"projects":   projectCostCentres,
		"others":     otherCostCentres,
	}); err != nil {
		return fmt.Errorf("Could not render template: %w", err)
	}
	return nil
}

func costCentrePage(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	costCentreIDString := r.PathValue("costCentreIDPath")
	costCentreIDInt, err := strconv.Atoi(costCentreIDString)
	if err != nil {
		return fmt.Errorf("failed to convert cost centre id from string to int: %v", err)
	}

	budgetLines, err := getBudgetLines(db, costCentreIDInt)
	if err != nil {
		return fmt.Errorf("failed get scan budget line information from database: %v", err)
	}

	//omg
	secondaryCostCentresWithBudgetLinesList := make([]secondaryCostCentresWithBudgetLines, 1)
	currentSecondaryCostCentre := &secondaryCostCentresWithBudgetLinesList[0]
	for _, budgetLine := range budgetLines {
		if currentSecondaryCostCentre.SecondaryCostCentreName != budgetLine.SecondaryCostCentreName {
			secondaryCostCentresWithBudgetLinesList = append(secondaryCostCentresWithBudgetLinesList, secondaryCostCentresWithBudgetLines{
				SecondaryCostCentreName: budgetLine.SecondaryCostCentreName,
				BudgetLines:             []excel.BudgetLine{},
			})
			currentSecondaryCostCentre = &secondaryCostCentresWithBudgetLinesList[len(secondaryCostCentresWithBudgetLinesList)-1]
		}
		currentSecondaryCostCentre.BudgetLines = append(currentSecondaryCostCentre.BudgetLines, budgetLine)
	}
	secondaryCostCentresWithBudgetLinesList = secondaryCostCentresWithBudgetLinesList[1:]

	costCentre, err := getCostCentreByID(db, costCentreIDInt)
	if err != nil {
		return fmt.Errorf("failed get scan cost centre information from database: %v", err)
	}

	if err := templates.ExecuteTemplate(w, "costcentre.html", map[string]any{
		"secondaryCostCentresWithBudgetLinesList": secondaryCostCentresWithBudgetLinesList,
		"costCentre": costCentre,
	}); err != nil {
		return fmt.Errorf("could not render template: %w", err)
	}
	return nil
}

type secondaryCostCentresWithBudgetLines struct {
	SecondaryCostCentreName string
	BudgetLines             []excel.BudgetLine
}

func getBudgetLines(db *sql.DB, costCentreID int) ([]excel.BudgetLine, error) {
	var budgetLinesGetStatementStatic = `
		SELECT 
    		budget_lines.id,
    		budget_lines.name,
    		income,
			expense,
			comment,
			account,
			secondary_cost_centres.id,
			secondary_cost_centres.name
		FROM budget_lines
		JOIN secondary_cost_centres ON secondary_cost_centres.id = secondary_cost_centre_id
		WHERE cost_centre_id = $1
		ORDER BY secondary_cost_centre_id
	`
	result, err := db.Query(budgetLinesGetStatementStatic, costCentreID)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget lines from database: %v", err)
	}
	var budgetLines []excel.BudgetLine
	for result.Next() {
		var budgetLine excel.BudgetLine

		err := result.Scan(
			&budgetLine.BudgetLineID,
			&budgetLine.BudgetLineName,
			&budgetLine.BudgetLineIncome,
			&budgetLine.BudgetLineExpense,
			&budgetLine.BudgetLineComment,
			&budgetLine.BudgetLineAccount,
			&budgetLine.SecondaryCostCentreID,
			&budgetLine.SecondaryCostCentreName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget line from query result: %v", err)
		}
		budgetLines = append(budgetLines, budgetLine)
	}
	return budgetLines, nil
}

func getCostCentres(db *sql.DB) ([]excel.CostCentre, error) {
	var costCentresGetStatementStatic = `SELECT id, name, type FROM cost_centres`
	result, err := db.Query(costCentresGetStatementStatic)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost centres from database: %v", err)
	}
	var costCentres []excel.CostCentre
	for result.Next() {
		var costCentre excel.CostCentre

		err := result.Scan(&costCentre.CostCentreID, &costCentre.CostCentreName, &costCentre.CostCentreType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost centre from query result: %v", err)
		}
		costCentres = append(costCentres, costCentre)
	}
	return costCentres, nil
}

func getCostCentreByID(db *sql.DB, costCentreID int) (excel.CostCentre, error) {
	var costCentreGetStatementStatic = `SELECT id, name, type FROM cost_centres WHERE id = $1`
	result := db.QueryRow(costCentreGetStatementStatic, costCentreID)
	var costCentre excel.CostCentre
	err := result.Scan(&costCentre.CostCentreID, &costCentre.CostCentreName, &costCentre.CostCentreType)
	if err != nil {
		return excel.CostCentre{}, fmt.Errorf("failed to scan cost centre from query result: %v", err)
	}
	return costCentre, nil
}

func splitCostCentresOnType(costCentres []excel.CostCentre) ([]excel.CostCentre, []excel.CostCentre, []excel.CostCentre, error) {
	var committeeCostCentres []excel.CostCentre
	var projectCostCentres []excel.CostCentre
	var otherCostCentres []excel.CostCentre
	for _, costCentre := range costCentres {
		switch costCentre.CostCentreType {
		case "committee":
			committeeCostCentres = append(committeeCostCentres, costCentre)
		case "project":
			projectCostCentres = append(projectCostCentres, costCentre)
		case "other":
			otherCostCentres = append(otherCostCentres, costCentre)
		default:
			return nil, nil, nil, fmt.Errorf("faulty cost centre type found when splitting")
		}
	}
	return committeeCostCentres, projectCostCentres, otherCostCentres, nil
}
