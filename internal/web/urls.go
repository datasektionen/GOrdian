package web

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/datasektionen/GOrdian/internal/config"
	"github.com/datasektionen/GOrdian/internal/database"
	"github.com/datasektionen/GOrdian/internal/excel"
	"html/template"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
)

type FrameLine struct {
	FrameLineName     string
	FrameLineIncome   int
	FrameLineExpense  int
	FrameLineInternal int
	FrameLineResult   int
}

const (
	loginSessionCookieName = "login-session"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFiles embed.FS

var templates *template.Template

func Mount(mux *http.ServeMux, db *sql.DB) error {
	var err error
	tokenURL := config.GetEnv().LoginURL + "/login?callback=" + config.GetEnv().ServerURL + "/token?token="
	templates, err = template.New("").Funcs(map[string]any{"formatMoney": formatMoney, "add": add, "sliceContains": sliceContains}).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}
	mux.Handle("/static/", http.FileServerFS(staticFiles))
	mux.Handle("/{$}", authRoute(db, indexPage, []string{}))
	mux.Handle("/costcentre/{costCentreIDPath}", authRoute(db, costCentrePage, []string{}))
	mux.Handle("/login", http.RedirectHandler(tokenURL, http.StatusSeeOther))
	mux.Handle("/token", route(db, tokenPage))
	mux.Handle("/logout", route(db, logoutPage))
	mux.Handle("/admin", authRoute(db, adminPage, []string{"admin", "view-all"}))
	mux.Handle("/admin/upload", authRoute(db, uploadPage, []string{"admin"}))
	mux.Handle("/api/CostCentres", cors(route(db, apiCostCentres)))
	mux.Handle("/api/SecondaryCostCentres", cors(route(db, apiSecondaryCostCentre)))
	mux.Handle("/api/BudgetLines", cors(route(db, apiBudgetLine)))
	mux.Handle("/framebudget", authRoute(db, framePage, []string{}))

	return nil
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
}

func add(x int, y int) int {
	return x + y
}

func formatMoney(value int) string {
	numStr := strconv.Itoa(value)
	length := len(numStr)
	var result string

	for i := 0; i < length; i++ {
		if i > 0 && (length-i)%3 == 0 {
			result += " "
		}
		result += string(numStr[i])
	}

	return result
}

func route(db *sql.DB, handler func(w http.ResponseWriter, r *http.Request, db *sql.DB) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r, db)
		if err != nil {
			slog.Error("Error from handler", "error", err)
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
		}
	})
}

func authRoute(db *sql.DB, handler func(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error, requiredPerms []string) http.Handler {
	return route(db, func(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
		loginCookie, err := r.Cookie(loginSessionCookieName)
		if err != nil {
			if len(requiredPerms) == 0 {
				return handler(w, r, db, []string{}, false)
			}
			slog.Error("failed to get login cookie", "error", err)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Not logged in"))
			return nil
		}
		loginUser, err := http.Get(config.GetEnv().LoginURL + "/verify/" + loginCookie.Value + "?api_key=" + config.GetEnv().LoginToken)
		if err != nil {
			return fmt.Errorf("no response from login: %v", err)
		}

		if loginUser.StatusCode != 200 {
			return handler(w, r, db, []string{}, false)
		}
		var loginBody struct {
			User string `json:"user"`
		}
		err = json.NewDecoder(loginUser.Body).Decode(&loginBody)
		if err != nil {
			return fmt.Errorf("failed to decode user body from json: %v", err)
		}
		userPerms, err := http.Get(config.GetEnv().PlsURL + "/api/user/" + loginBody.User + "/" + config.GetEnv().PlsSystem)

		var perms []string
		err = json.NewDecoder(userPerms.Body).Decode(&perms)
		if err != nil {
			return fmt.Errorf("failed to decode perms body from json: %v", err)
		}

		if !sliceContains(requiredPerms, perms...) && len(requiredPerms) != 0 {
			slog.Error("Error from handler", "error", err)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return nil
		}
		return handler(w, r, db, perms, true)
	})

}

func sliceContains(list1 []string, list2 ...string) bool {
	// Iterate through list1 and check if one object is present in list2
	for _, obj1 := range list1 {
		for _, obj2 := range list2 {
			if obj1 == obj2 {
				return true
			}
		}
	}
	return false
}

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

func adminPage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {
	if err := templates.ExecuteTemplate(w, "admin.html", map[string]any{
		"motd":        motdGenerator(),
		"permissions": perms,
		"loggedIn":    loggedIn,
	}); err != nil {
		return fmt.Errorf("could not render template: %w", err)
	}
	return nil
}

func uploadPage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {
	file, _, err := r.FormFile("budgetFile")
	if err != nil {
		return fmt.Errorf("could not read file from form: %w", err)
	}
	err = database.SaveBudget(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
	return nil
}

func tokenPage(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	sessionCookieVal := r.FormValue("token")
	sessionCookie := http.Cookie{Name: loginSessionCookieName, Value: sessionCookieVal}
	http.SetCookie(w, &sessionCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func logoutPage(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	sessionCookie := http.Cookie{Name: loginSessionCookieName, MaxAge: -1}
	http.SetCookie(w, &sessionCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func indexPage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {
	costCentres, err := getCostCentres(db)
	if err != nil {
		return fmt.Errorf("failed get scan cost centre information from database: %v", err)
	}
	committeeCostCentres, projectCostCentres, otherCostCentres, err := splitCostCentresOnType(costCentres)
	if err != nil {
		return fmt.Errorf("failed to split cost centres on type: %v", err)
	}

	//Mörkläggning av mottagningens budget
	darkeningResp, err := http.Get("https://cashflow.datasektionen.se/")
	if err != nil {
		slog.Error("Failed to get status from darkmode", "error", err)
		return fmt.Errorf(": %v", err)
	}
	defer darkeningResp.Body.Close()

	if darkeningResp.StatusCode != http.StatusOK {
		slog.Error("Status error from darkmode", "error", darkeningResp.StatusCode)
	}

	darkeningBody, err := io.ReadAll(darkeningResp.Body)
	if err != nil {
		slog.Error("Failed to read body", "error", err)
	}

	darkeningValue, err := strconv.ParseBool(string(darkeningBody))
	if err != nil {
		slog.Error("Failed to parse bool", "error", err)
	}

	if darkeningValue {
		for index, committeeCostCentre := range committeeCostCentres {
			if committeeCostCentre.CostCentreName == "Mottagningen" {
				committeeCostCentres = append(committeeCostCentres[:index], committeeCostCentres[index+1:]...)
				break
			}
		}
	}
	//end of mörkläggning

	if err := templates.ExecuteTemplate(w, "index.html", map[string]any{
		"motd":        motdGenerator(),
		"committees":  committeeCostCentres,
		"projects":    projectCostCentres,
		"others":      otherCostCentres,
		"permissions": perms,
		"loggedIn":    loggedIn,
	}); err != nil {
		return fmt.Errorf("Could not render template: %w", err)
	}
	return nil
}

func framePage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {
	budgetLines, err := getFrameLines(db)
	if err != nil {
		return fmt.Errorf("failed get scan budget lines information from database: %v", err)
	}
	committeeFrameLines, projectFrameLines, otherFrameLines, totalFrameLine, err := generateFrameLines(budgetLines)
	if err != nil {
		return fmt.Errorf("failed to generate frame budget lines: %v", err)
	}
	if err := templates.ExecuteTemplate(w, "frame.html", map[string]any{
		"motd":                motdGenerator(),
		"committeeframelines": committeeFrameLines,
		"projectframelines":   projectFrameLines,
		"otherframelines":     otherFrameLines,
		"totalframeline":      totalFrameLine,
		"permissions":         perms,
		"loggedIn":            loggedIn,
	}); err != nil {
		return fmt.Errorf("Could not render template: %w", err)
	}
	return nil
}

func costCentrePage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {
	costCentreIDString := r.PathValue("costCentreIDPath")
	costCentreIDInt, err := strconv.Atoi(costCentreIDString)
	if err != nil {
		return fmt.Errorf("failed to convert cost centre id from string to int: %v", err)
	}

	budgetLines, err := getBudgetLinesByCostCentreID(db, costCentreIDInt)
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

	//calc the total incomes, expenses and results of all cost centres in the list
	secondaryCostCentresWithBudgetLinesList, err = calculateSecondaryCostCentres(secondaryCostCentresWithBudgetLinesList)
	if err != nil {
		return fmt.Errorf("failed calculate secondary cost centre values: %v", err)
	}

	costCentreTotalIncome, costCentreTotalExpense, costCentreTotalResult, err := calculateCostCentre(secondaryCostCentresWithBudgetLinesList)
	if err != nil {
		return fmt.Errorf("failed calculate cost centre values: %v", err)
	}

	if err := templates.ExecuteTemplate(w, "costcentre.html", map[string]any{
		"motd": motdGenerator(),
		"secondaryCostCentresWithBudgetLinesList": secondaryCostCentresWithBudgetLinesList,
		"costCentre":             costCentre,
		"costCentreTotalIncome":  costCentreTotalIncome,
		"costCentreTotalExpense": costCentreTotalExpense,
		"costCentreTotalResult":  costCentreTotalResult,
		"permissions":            perms,
		"loggedIn":               loggedIn,
	}); err != nil {
		return fmt.Errorf("could not render template: %w", err)
	}
	return nil
}

func calculateCostCentre(secondaryCostCentresWithBudgetLinesList []secondaryCostCentresWithBudgetLines) (int, int, int, error) {
	var totalIncome int
	var totalExpense int
	for _, sCCWithBudgetLines := range secondaryCostCentresWithBudgetLinesList {
		totalIncome = totalIncome + sCCWithBudgetLines.SecondaryCostCentreTotalIncome
		totalExpense = totalExpense + sCCWithBudgetLines.SecondaryCostCentreTotalExpense
	}
	totalResult := totalIncome + totalExpense

	return totalIncome, totalExpense, totalResult, nil
}

func calculateSecondaryCostCentres(secondaryCostCentresWithBudgetLinesList []secondaryCostCentresWithBudgetLines) ([]secondaryCostCentresWithBudgetLines, error) {
	for index, sCCWithBudgetLines := range secondaryCostCentresWithBudgetLinesList {
		var totalIncome int
		var totalExpense int
		for _, budgetLine := range sCCWithBudgetLines.BudgetLines {
			totalIncome = totalIncome + budgetLine.BudgetLineIncome
			totalExpense = totalExpense + budgetLine.BudgetLineExpense
		}
		secondaryCostCentresWithBudgetLinesList[index].SecondaryCostCentreTotalIncome = totalIncome
		secondaryCostCentresWithBudgetLinesList[index].SecondaryCostCentreTotalExpense = totalExpense
		secondaryCostCentresWithBudgetLinesList[index].SecondaryCostCentreTotalResult = totalIncome + totalExpense
	}
	return secondaryCostCentresWithBudgetLinesList, nil
}

type secondaryCostCentresWithBudgetLines struct {
	SecondaryCostCentreName         string
	SecondaryCostCentreTotalIncome  int
	SecondaryCostCentreTotalExpense int
	SecondaryCostCentreTotalResult  int
	BudgetLines                     []excel.BudgetLine
}

func getBudgetLinesByCostCentreID(db *sql.DB, costCentreID int) ([]excel.BudgetLine, error) {
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

func getSecondaryCostCentresByCostCentreID(db *sql.DB, costCentreID int) ([]excel.SecondaryCostCentre, error) {
	var SecondaryCostCentresGetStatementStatic = `
		SELECT 
    		id,
    		name,
    		cost_centre_id
		FROM secondary_cost_centres
		WHERE cost_centre_id = $1
		ORDER BY id
	`
	result, err := db.Query(SecondaryCostCentresGetStatementStatic, costCentreID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secondary cost centres from database: %v", err)
	}
	var secondaryCostCentres []excel.SecondaryCostCentre
	for result.Next() {
		var secondaryCostCentre excel.SecondaryCostCentre

		err := result.Scan(
			&secondaryCostCentre.SecondaryCostCentreID,
			&secondaryCostCentre.SecondaryCostCentreName,
			&secondaryCostCentre.CostCentreID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secondary cost centre from query result: %v", err)
		}
		secondaryCostCentres = append(secondaryCostCentres, secondaryCostCentre)
	}
	return secondaryCostCentres, nil
}

func getBudgetLinesBySecondaryCostCentreID(db *sql.DB, secondaryCostCentreID int) ([]excel.BudgetLine, error) {
	var budgetLinesGetStatementStatic = `
		SELECT 
    		id,
    		name,
    		income,
			expense,
			comment,
			account,
			secondary_cost_centre_id
		FROM budget_lines
		WHERE secondary_cost_centre_id = $1
		ORDER BY id
	`
	result, err := db.Query(budgetLinesGetStatementStatic, secondaryCostCentreID)
	if err != nil {
		return nil, fmt.Errorf("failed to get budgetlines from database: %v", err)
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget line from query result: %v", err)
		}
		budgetLines = append(budgetLines, budgetLine)
	}
	return budgetLines, nil
}

func getFrameLines(db *sql.DB) ([]excel.BudgetLine, error) {
	var frameLinesGetStatementStatic = `
		SELECT 
    		SUM(income),
			SUM(expense),
			secondary_cost_centres.name = 'Internt',
			cost_centres.id,
			cost_centres.name,
			cost_centres.type
		FROM budget_lines
		JOIN secondary_cost_centres ON secondary_cost_centres.id = secondary_cost_centre_id
		JOIN cost_centres ON secondary_cost_centres.cost_centre_id = cost_centres.id
		GROUP BY cost_centres.id, cost_centres.name, cost_centres.type, secondary_cost_centres.name = 'Internt'
		ORDER BY cost_centres.name, secondary_cost_centres.name = 'Internt'
	`
	result, err := db.Query(frameLinesGetStatementStatic)
	if err != nil {
		return nil, fmt.Errorf("failed to get framelines from database: %v", err)
	}
	var frameLines []excel.BudgetLine
	for result.Next() {
		var frameLine excel.BudgetLine

		err := result.Scan(
			&frameLine.BudgetLineIncome,
			&frameLine.BudgetLineExpense,
			&frameLine.SecondaryCostCentreName,
			&frameLine.CostCentreID,
			&frameLine.CostCentreName,
			&frameLine.CostCentreType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget line from query result: %v", err)
		}
		frameLines = append(frameLines, frameLine)
	}
	return frameLines, nil
}

func getCostCentres(db *sql.DB) ([]excel.CostCentre, error) {
	var costCentresGetStatementStatic = `SELECT id, name, type FROM cost_centres ORDER BY name`
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

func generateFrameLines(frameLines []excel.BudgetLine) ([]FrameLine, []FrameLine, []FrameLine, FrameLine, error) {
	var committeeFrameLines []FrameLine
	var projectFrameLines []FrameLine
	var otherFrameLines []FrameLine
	var totalFrameLine FrameLine
	totalFrameLine.FrameLineName = "Totalt"
	var skippidi bool
	for i, frameLine := range frameLines {
		if skippidi == true {
			skippidi = false
			continue
		}
		frameLineIncome := frameLine.BudgetLineIncome
		frameLineExpense := frameLine.BudgetLineExpense
		frameLineName := frameLine.CostCentreName
		frameLineInternal := 0
		frameLineResult := 0

		if i+1 < len(frameLines) && frameLines[i+1].CostCentreName == frameLineName {
			frameLineIncome += frameLines[i+1].BudgetLineIncome
			frameLineExpense += frameLines[i+1].BudgetLineExpense
			frameLineInternal = frameLines[i+1].BudgetLineIncome + frameLines[i+1].BudgetLineExpense
			skippidi = true
		}

		frameLineResult = frameLineIncome + frameLineExpense

		reconstructedFrameLine := FrameLine{frameLineName, frameLineIncome, frameLineExpense, frameLineInternal, frameLineResult}

		totalFrameLine.FrameLineIncome += frameLineIncome
		totalFrameLine.FrameLineExpense += frameLineExpense
		totalFrameLine.FrameLineInternal += frameLineInternal
		totalFrameLine.FrameLineResult += frameLineResult

		switch frameLine.CostCentreType {
		case "committee":
			committeeFrameLines = append(committeeFrameLines, reconstructedFrameLine)
		case "project":
			projectFrameLines = append(projectFrameLines, reconstructedFrameLine)
		case "other":
			otherFrameLines = append(otherFrameLines, reconstructedFrameLine)
		default:
			return nil, nil, nil, FrameLine{}, fmt.Errorf("faulty cost centre type found when splitting")
		}
	}
	return committeeFrameLines, projectFrameLines, otherFrameLines, totalFrameLine, nil
}

func motdGenerator() string {
	options := []string{
		"You have very many money:",
		"Sjunde gången gillt:",
		"Kassörens bästa vän:",
		"Brought to you by FIPL consulting:",
		"Kom på hackerkvällarna!",
		"12345690,+"}
	randomIndex := rand.Intn(len(options))
	return options[randomIndex]
}
