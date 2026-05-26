package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"golang.org/x/crypto/bcrypt"
)

// cookieName is the session cookie identifier
const cookieName = "drukarena_session"
const adminEmail = "drukarena@gmail.com"
const adminPassword = "admin@123"

// sessionStore is an in-memory map: cookie value → user_id
// In production you'd use Redis or a DB-backed session table
var sessionStore = map[string]int64{}

// Signup handles player registration.
func Signup(w http.ResponseWriter, r *http.Request) {
	var u model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()

	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	if err := validateSignupInput(u.Username, u.Email, u.Password); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Hash password with bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, "could not hash password")
		return
	}
	u.Password = string(hashed)

	u.Role = "user"

	if err := u.Create(); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "user already exists or invalid data: "+err.Error())
		return
	}
	_ = model.EnsurePlayerProfile(u.UserID, u.Username)

	token := generateToken(u.UserID)
	sessionStore[token] = u.UserID
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  time.Now().Add(30 * time.Minute),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":   "account created",
		"user_id":  u.UserID,
		"role":     u.Role,
		"username": u.Username,
	})
}

// Login authenticates a user and sets a session cookie (30-min expiry)
func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&creds); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	creds.Email = strings.TrimSpace(strings.ToLower(creds.Email))
	if creds.Email == "" || creds.Password == "" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if _, err := mail.ParseAddress(creds.Email); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "enter a valid email address")
		return
	}

	if creds.Email == adminEmail && creds.Password == adminPassword {
		hashed, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			httpResp.RespondWithError(w, http.StatusInternalServerError, "could not prepare admin account")
			return
		}
		if err := model.EnsureAdminUser(adminEmail, string(hashed)); err != nil {
			httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	u := model.User{Email: creds.Email}
	if err := u.GetByEmail(); err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "Unauthorized. Credentials does not match!")
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(creds.Password)); err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "Unauthorized. Credentials does not match!")
		return
	}
	if u.Role == "admin" && (u.Email != adminEmail || creds.Password != adminPassword) {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin dashboard requires DrukArena admin credentials")
		return
	}
	_ = model.EnsurePlayerProfile(u.UserID, u.Username)

	// Create session token (simple unique string)
	token := generateToken(u.UserID)
	sessionStore[token] = u.UserID

	// Set cookie — 30 min expiry following CSC103 pattern
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  time.Now().Add(30 * time.Minute),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	u.Password = "" // never send password back
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "login successful",
		"user_id":  u.UserID,
		"role":     u.Role,
		"username": u.Username,
	})
}

// Logout clears the session cookie
func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		delete(sessionStore, cookie.Value)
	}
	// Delete cookie at client side by setting MaxAge=-1
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// VerifySession checks if the session cookie is valid and returns user info
func VerifySession(w http.ResponseWriter, r *http.Request) {
	userID, role, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"role":    role,
		"username": func() string {
			u := model.User{UserID: userID}
			if err := u.GetByID(); err == nil {
				return u.Username
			}
			return ""
		}(),
	})
}

// GetSessionUser extracts user_id and role from the request cookie — used by all protected handlers
func GetSessionUser(r *http.Request) (int64, string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return 0, "", err
	}
	userID, ok := sessionStore[cookie.Value]
	if !ok {
		return 0, "", http.ErrNoCookie
	}
	u := model.User{UserID: userID}
	if err := u.GetByID(); err != nil {
		return 0, "", err
	}
	return u.UserID, u.Role, nil
}

// generateToken creates a simple unique session token
func generateToken(userID int64) string {
	return fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
}

func validateSignupInput(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return errString("username, email and password are required")
	}
	if len(username) < 3 {
		return errString("username must be at least 3 characters")
	}
	if len(username) > 40 {
		return errString("username must be 40 characters or less")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errString("enter a valid email address")
	}
	if len(password) < 8 {
		return errString("password must be at least 8 characters")
	}
	hasLetter, hasNumber := false, false
	for _, ch := range password {
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' {
			hasLetter = true
		}
		if ch >= '0' && ch <= '9' {
			hasNumber = true
		}
	}
	if !hasLetter || !hasNumber {
		return errString("password must include at least one letter and one number")
	}
	return nil
}
