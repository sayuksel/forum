package handlers
import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"forum/errors"
	"html/template"
	"net/http"
	"strings"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)
var db *sql.DB
// Initialize database
func InitDB() {
	var err error
	db, err = sql.Open("sqlite", "forum.db") // Use "sqlite" instead of "sqlite3"
	if err != nil {
		panic(err)
	}
}
func GetDB() *sql.DB {
	return db
}
var temp = template.Must(template.ParseGlob("templates/*.html"))
// Register handler
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		// Check if email or username is already taken
		var existingUser int
		err := db.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", email, username).Scan(&existingUser)
		if err == nil {
			errors.Error409(w, r)
			return
		}
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		// Insert the new user into the database
		_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, string(hashedPassword))
		if err != nil {
			errors.Error500(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Render the registration form
	temp.ExecuteTemplate(w, "register.html", nil)
}
var sessionStore = map[string]int{} // Simple in-memory session store: session_token -> userID
func generateSessionToken() (string, error) {
	b := make([]byte, 32) // Generate 32 random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Return the random bytes encoded in base64 format (URL-safe encoding)
	return base64.URLEncoding.EncodeToString(b), nil
}
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		email := strings.TrimSpace(r.FormValue("email"))
		password := strings.TrimSpace(r.FormValue("password"))
		var storedPassword string
		var userID int
		// Query the database for the user
		err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &storedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				errors.Error401(w, r)
				return
			}
			errors.Error500(w, r)
			return
		}
		// Compare the hashed password
		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			errors.Error401(w, r)
			return
		}
		// Invalidate any existing session for this user
		for token, id := range sessionStore {
			if id == userID {
				delete(sessionStore, token) // Remove the old session
				break
			}
		}
		// Generate a new session token
		sessionToken, err := generateSessionToken()
		if err != nil {
			errors.Error500(w, r)
			return
		}
		// Store the new session token in the session store
		sessionStore[sessionToken] = userID
		// Set the new session token as a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Path:     "/",
			MaxAge:   3600, // 1-hour expiry
			HttpOnly: true,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// Render the login form if not a POST request
	err := temp.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		errors.Error500(w, r)
	}
}
func Logout(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" || r.Method == "GET" { // Allow GET method for logout
        // Clear the session cookie
        http.SetCookie(w, &http.Cookie{
            Name:     "session_token",
            Value:    "",
            Path:     "/",
            MaxAge:   -1, // Invalidate the cookie
            HttpOnly: true,
        })
        // Remove the session token from the session store
        cookie, err := r.Cookie("session_token")
        if err == nil {
            delete(sessionStore, cookie.Value)
        }
        // Redirect to the home page
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

