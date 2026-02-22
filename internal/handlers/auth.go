package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lukas-pastva/web-charon/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userContextKey contextKey = "user"

type AuthHandler struct {
	Users         *models.UserStore
	Templates     *template.Template
	SessionSecret []byte
}

// CurrentUser extracts the authenticated user from request context.
func CurrentUser(r *http.Request) *models.User {
	u, _ := r.Context().Value(userContextKey).(*models.User)
	return u
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.renderLogin(w, "")
}

func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	nickname := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.Users.GetByNickname(nickname)
	if err != nil {
		h.renderLogin(w, "Invalid username or password.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		h.renderLogin(w, "Invalid username or password.")
		return
	}

	value := fmt.Sprintf("%d:%s", user.ID, time.Now().Format(time.RFC3339))
	sig := h.sign(value)
	cookie := &http.Cookie{
		Name:     "charon_session",
		Value:    value + "|" + sig,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7,
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
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		user := h.userFromSession(cookie.Value)
		if user == nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := CurrentUser(r)
		if user == nil || !user.IsAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AuthHandler) userFromSession(raw string) *models.User {
	// Format: "userID:timestamp|signature"
	idx := strings.LastIndex(raw, "|")
	if idx < 0 {
		return nil
	}
	value := raw[:idx]
	sig := raw[idx+1:]
	if sig != h.sign(value) {
		return nil
	}

	// Extract user ID from "id:timestamp"
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return nil
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil
	}

	user, err := h.Users.GetByID(userID)
	if err != nil {
		return nil
	}
	return user
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
