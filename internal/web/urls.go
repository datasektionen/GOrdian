package web

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/datasektionen/GOrdian/internal/config"
	"github.com/datasektionen/GOrdian/internal/database"
)

type Databases struct {
	DBCF *sql.DB
	DBGO *sql.DB
}

const (
	loginSessionCookieName = "login-session"
)

//go:embed templates/*.gohtml
var templatesFS embed.FS

//go:embed static/*
var staticFiles embed.FS

var templates *template.Template

func Mount(mux *http.ServeMux, databases Databases) error {
	var err error
	tokenURL := config.GetEnv().LoginURL + "/login?callback=" + config.GetEnv().ServerURL + "/token?token="
	templates, err = template.New("").Funcs(map[string]any{"formatMoney": formatMoney, "add": add, "sliceContains": sliceContains}).ParseFS(templatesFS, "templates/*.gohtml")
	if err != nil {
		return err
	}
	mux.Handle("/static/", http.FileServerFS(staticFiles))
	mux.Handle("/{$}", authRoute(databases.DBGO, indexPage, []string{}))
	mux.Handle("/costcentre/{costCentreIDPath}", authRoute(databases.DBGO, costCentrePage, []string{}))
	mux.Handle("/login", http.RedirectHandler(tokenURL, http.StatusSeeOther))
	mux.Handle("/token", route(databases.DBGO, tokenPage))
	mux.Handle("/logout", route(databases.DBGO, logoutPage))
	mux.Handle("/admin", authRoute(databases.DBGO, adminPage, []string{"admin", "view-all"}))
	mux.Handle("/admin/upload", authRoute(databases.DBGO, uploadPage, []string{"admin"}))
	mux.Handle("/api/CostCentres", cors(route(databases.DBGO, apiCostCentres)))
	mux.Handle("/api/SecondaryCostCentres", cors(route(databases.DBGO, apiSecondaryCostCentre)))
	mux.Handle("/api/BudgetLines", cors(route(databases.DBGO, apiBudgetLine)))
	mux.Handle("/framebudget", authRoute(databases.DBGO, framePage, []string{}))
	mux.Handle("/resultreport", authRoute(databases.DBCF, reportPage, []string{}))

	return nil
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
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
		if err != nil {
			return fmt.Errorf("no response from pls: %v", err)
		}

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

func adminPage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {
	if err := templates.ExecuteTemplate(w, "admin.gohtml", map[string]any{
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
	err = database.SaveBudget(file, db)
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
