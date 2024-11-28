package models
import "time"
type Posts struct {
	ID         int
	User_ID    int
	Username   string
	Title      string
	Content    string
	Category   string
	CreatedAt time.Time
	LikesCount int
	DislikesCount int
}
