package handlers

import (
	"crypto/subtle"
	"net/http"
	"time"
)

// LoginData holds template data for the login page.
type LoginData struct {
	Error string
}

// LoginPage renders the login form.
func (a *App) LoginPage(w http.ResponseWriter, r *http.Request) {
	if a.Auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	a.render(w, "login", LoginData{})
}

// LoginHandler processes login form submissions.
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	if subtle.ConstantTimeCompare([]byte(password), []byte(a.Config.AdminPassword)) != 1 {
		a.render(w, "login", LoginData{Error: "Invalid password"})
		return
	}

	token := a.Auth.GenerateToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "fileorbit_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7, // 7 days
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// LogoutHandler clears the session cookie and redirects to login.
func (a *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "fileorbit_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RootHandler redirects to dashboard if authenticated, otherwise to login.
func (a *App) RootHandler(w http.ResponseWriter, r *http.Request) {
	if a.Auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
