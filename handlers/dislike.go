// Handle dislike button submissions for posts
package handlers
import (
	"database/sql"
	"fmt"
	"forum/errors"
	"net/http"
)
// Handle dislike button submissions for posts
func DislikePost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	var currentDislike bool
	err = db.QueryRow("SELECT dislike FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&currentDislike)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO post_likes (user_id, post_id, like, dislike) VALUES (?, ?, FALSE, TRUE)", userID, postID)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		_, err = db.Exec("UPDATE posts SET dislikes_count = dislikes_count + 1 WHERE id = ?", postID)
	} else if currentDislike {
		_, err = db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		_, err = db.Exec("UPDATE posts SET dislikes_count = dislikes_count - 1 WHERE id = ?", postID)
	} else {
		_, err = db.Exec("UPDATE post_likes SET like = FALSE, dislike = TRUE WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		_, err = db.Exec("UPDATE posts SET dislikes_count = dislikes_count + 1, likes_count = likes_count - 1 WHERE id = ?", postID)
	}
	if err != nil {
		errors.Error500(w, r)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/postpage?id=%s", postID), http.StatusSeeOther)
}
// Handle dislike button submissions for comments
func DislikeComment(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	var currentDislike bool
	err = db.QueryRow("SELECT dislike FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&currentDislike)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO comment_likes (user_id, comment_id, like, dislike) VALUES (?, ?, FALSE, TRUE)", userID, commentID)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		_, err = db.Exec("UPDATE comments SET dislikes_count = dislikes_count + 1 WHERE id = ?", commentID)
	} else if currentDislike {
		_, err = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		_, err = db.Exec("UPDATE comments SET dislikes_count = dislikes_count - 1 WHERE id = ?", commentID)
	} else {
		_, err = db.Exec("UPDATE comment_likes SET like = FALSE, dislike = TRUE WHERE user_id = ? AND comment_id = ?", userID, commentID)
		if err != nil {
			errors.Error500(w, r)
			return
		}
		_, err = db.Exec("UPDATE comments SET dislikes_count = dislikes_count + 1, likes_count = likes_count - 1 WHERE id = ?", commentID)
	}
	if err != nil {
		errors.Error500(w, r)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/postpage?id=%s", postID), http.StatusSeeOther)
}
