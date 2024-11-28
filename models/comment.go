
package models
import "time"
type Comment struct {
    ID            int
    UserID        int    // Camel case
    Username      string
    PostID        int    // Camel case
    Content       string
    CreatedAt     time.Time
    LikesCount    int
    DislikesCount int
}
