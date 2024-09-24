package web

import (
	"database/sql"
	"fmt"
	"net/http"
)

func reportPage(w http.ResponseWriter, r *http.Request, db *sql.DB, perms []string, loggedIn bool) error {

	if err := templates.ExecuteTemplate(w, "report.gohtml", map[string]any{
		"motd":        motdGenerator(),
		"permissions": perms,
		"loggedIn":    loggedIn,
	}); err != nil {
		return fmt.Errorf("Could not render template: %w", err)
	}
	return nil
}
