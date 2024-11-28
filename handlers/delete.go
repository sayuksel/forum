package handlers
import (
    "database/sql"
    "fmt"
    "net/http"
    "forum/errors"
)
// Handle delete button submissions for posts
func DeletePost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        errors.Error405(w, r)
        return
    }
    if err := r.ParseForm(); err != nil {
        errors.Error400(w, r)
        return
    }
    postID := r.FormValue("post_id")
    if postID == "" {
        errors.Error400(w, r)
        return
    }
    cookie, err := r.Cookie("session_token")
    if err != nil {
        errors.Error401(w, r)
        return
    }
    userID, ok := sessionStore[cookie.Value]
    if !ok {
        errors.Error401(w, r)
        return
    }
    // Check if the post belongs to the user
    var postOwnerID int
    err = db.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwnerID)
    if err != nil {
        errors.Error404(w, r)
        return
    }
    if postOwnerID != userID {
        errors.Error401(w, r)
        return
    }
    // Delete the post and its comments
    _, err = db.Exec("DELETE FROM comments WHERE post_id = ?", postID)
    if err != nil {
        errors.Error500(w, r)
        return
    }
    _, err = db.Exec("DELETE FROM posts WHERE id = ?", postID)
    if err != nil {
        errors.Error500(w, r)
        return
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
// Handle delete button submissions for comments
func DeleteComment(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        errors.Error405(w, r)
        return
    }
    if err := r.ParseForm(); err != nil {
        errors.Error400(w, r)
        return
    }
    commentID := r.FormValue("comment_id")
    postID := r.FormValue("post_id")
    if commentID == "" || postID == "" {
        errors.Error400(w, r)
        return
    }
    cookie, err := r.Cookie("session_token")
    if err != nil {
        errors.Error401(w, r)
        return
    }
    userID, ok := sessionStore[cookie.Value]
    if !ok {
        errors.Error401(w, r)
        return
    }
    // Check if the comment belongs to the user
    var commentOwnerID int
    err = db.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentOwnerID)
    if err != nil {
        errors.Error404(w, r)
        return
    }
    if commentOwnerID != userID {
        errors.Error401(w, r)
        return
    }
    // Delete the comment
    _, err = db.Exec("DELETE FROM comments WHERE id = ?", commentID)
    if err != nil {
        errors.Error500(w, r)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/postpage?id=%s", postID), http.StatusSeeOther)
}