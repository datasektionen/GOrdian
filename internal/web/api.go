package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func apiCostCentres(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	costCentres, err := getCostCentres(db)
	if err != nil {
		return fmt.Errorf("failed get scan cost centres information from database: %v", err)
	}
	err = json.NewEncoder(w).Encode(costCentres)
	if err != nil {
		return fmt.Errorf("failed to encode cost centres to json: %v", err)
	}
	return nil
}

func apiSecondaryCostCentre(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	idCC, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		return fmt.Errorf("failed to convert secondary cost centre id to int: %v", err)
	}
	secondaryCostCentres, err := getSecondaryCostCentresByCostCentreID(db, idCC)
	if err != nil {
		return fmt.Errorf("failed get scan sendondary cost centres information from database: %v", err)
	}
	err = json.NewEncoder(w).Encode(secondaryCostCentres)
	if err != nil {
		return fmt.Errorf("failed to encode secondary cost centres to json: %v", err)
	}
	return nil
}

func apiBudgetLine(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	idSCC, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		return fmt.Errorf("failed to convert SCC id fromstring to int: %v", err)
	}
	budgetLines, err := getBudgetLinesBySecondaryCostCentreID(db, idSCC)
	if err != nil {
		return fmt.Errorf("failed get scan budget lines information from database: %v", err)
	}
	err = json.NewEncoder(w).Encode(budgetLines)
	if err != nil {
		return fmt.Errorf("failed to encode budget lines to json: %v", err)
	}
	return nil
}
