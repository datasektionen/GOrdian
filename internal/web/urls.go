package web

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/datasektionen/GOrdian/internal/excel"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates *template.Template

func Mount(mux *http.ServeMux, db *sql.DB) error {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}

	mux.Handle("/{$}", page(db, index))

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

func getCostCentres(db *sql.DB) ([]excel.CostCentre, error) {
	var costCentresGetStatementStatic = `SELECT id, name, type FROM cost_centres`
	result, err := db.Query(costCentresGetStatementStatic)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost centre from database: %v", err)
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
