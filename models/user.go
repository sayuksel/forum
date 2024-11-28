// models/user.go
package models
import (
    "time"
)
type User struct {
    ID            int       `json:"id"`
    Username      string    `json:"username"`
    Email         string    `json:"email"`
    Password      string    `json:"-"` // Do not expose password
    SessionToken  string    `json:"session_token"`
    SessionExpiry time.Time `json:"session_expiry"`
}

