package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"time"
)

type AuthHandler struct {
	AdminPassword string
	Templates     *template.Template
	SessionSecret []byte
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.renderLogin(w, "")
}

func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != "admin" || password != h.AdminPassword {
		h.renderLogin(w, "Invalid username or password.")
		return
	}

	value := "admin:" + time.Now().Format(time.RFC3339)
	sig := h.sign(value)
	cookie := &http.Cookie{
		Name:     "charon_session",
		Value:    value + "|" + sig,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7, // 7 days
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "charon_session",
		Value:    "",
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (h *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("charon_session")
		if err != nil || !h.validSession(cookie.Value) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AuthHandler) validSession(raw string) bool {
	// Format: "value|signature"
	for i := len(raw) - 1; i >= 0; i-- {
		if raw[i] == '|' {
			value := raw[:i]
			sig := raw[i+1:]
			return sig == h.sign(value)
		}
	}
	return false
}

func (h *AuthHandler) sign(value string) string {
	mac := hmac.New(sha256.New, h.SessionSecret)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *AuthHandler) renderLogin(w http.ResponseWriter, errMsg string) {
	data := map[string]interface{}{
		"Error": errMsg,
	}
	err := h.Templates.ExecuteTemplate(w, "login.html", data)
	if err != nil {
		log.Printf("login template error: %v", err)
		http.Error(w, "Internal Server Error", 500)
	}
}
