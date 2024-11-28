package handlers
import (
	"database/sql"
	"fmt"
	"strings"
	"forum/errors"
	"forum/models"
	"log"
	"net/http"
	"time"
)
// leading to post page
// PostPage renders the detailed view of a single post.
func PostPage(w http.ResponseWriter, r *http.Request) {
    // Get the post ID from the query parameters
    postID := r.URL.Query().Get("id")
    if postID == "" {
        http.NotFound(w, r)
        return
    }
    // Fetch the post by ID
    var post models.Posts
    err := db.QueryRow(`
        SELECT p.id, p.user_id, u.username, p.title, p.content, IFNULL(GROUP_CONCAT(c.name), '') AS categories, p.created_at, p.likes_count, p.dislikes_count
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN post_categories pc ON p.id = pc.post_id
        LEFT JOIN categories c ON pc.category_id = c.id
        WHERE p.id = ?
        GROUP BY p.id`, postID).Scan(
        &post.ID,
        &post.User_ID,
        &post.Username,
        &post.Title,
        &post.Content,
        &post.Category,
        &post.CreatedAt,
        &post.LikesCount,
        &post.DislikesCount,
    )
    if err != nil {
        log.Println("Error fetching post:", err)
        http.NotFound(w, r)
        return
    }
    // Fetch comments for the post
    rows, err := db.Query(`
        SELECT c.id, c.user_id, c.content, c.created_at, u.username, c.likes_count, c.dislikes_count
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC`, postID)
    if err != nil {
        log.Println("Error fetching comments:", err)
        errors.Error500(w, r)
        return
    }
    defer rows.Close()
    var comments []models.Comment
    for rows.Next() {
        var comment models.Comment
        if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.Username, &comment.LikesCount, &comment.DislikesCount); err != nil {
            log.Println("Error scanning comment:", err)
            errors.Error500(w, r)
            return
        }
        comments = append(comments, comment)
    }
    // Get session details
    cookie, err := r.Cookie("session_token")
    var user models.User
    isAuthenticated := false
    if err == nil {
        // Validate session
        userID, ok := sessionStore[cookie.Value]
        if ok {
            isAuthenticated = true
            err := db.QueryRow("SELECT id, username FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username)
            if err != nil {
                log.Println("Error retrieving user data:", err)
            }
        }
    }
    // Prepare data for the template
    data := struct {
        User            models.User
        IsAuthenticated bool
        Post            models.Posts
        Comments        []models.Comment
    }{
        User:            user,
        IsAuthenticated: isAuthenticated,
        Post:            post,
        Comments:        comments,
    }
    // Render the postpage template
    if err := templates.ExecuteTemplate(w, "postpage.html", data); err != nil {
        log.Println("Error rendering post page template:", err)
        errors.Error500(w, r)
        return
    }
}
// fetches a specific post by ID to display on postpage
func fetchPostByID(id string) (models.Posts, error) {
    var post models.Posts
    // Query the db to get post information
    err := db.QueryRow(`
        SELECT p.id, p.user_id, u.username, p.title, p.content, p.category, p.created_at, p.likes_count, p.dislikes_count
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = ?`, id).
        Scan(&post.ID, &post.User_ID, &post.Username, &post.Title, &post.Content, &post.Category, &post.CreatedAt, &post.LikesCount, &post.DislikesCount)
    if err != nil {
        return post, err
    }
    return post, nil
}
// Fetch comments for a specific post
func fetchCommentsByPostID(postID string) ([]models.Comment, error) {
    rows, err := db.Query(`
        SELECT c.id, c.user_id, c.post_id, c.content, c.created_at, u.username, c.likes_count, c.dislikes_count
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at DESC`, postID)
    if err != nil {
        log.Printf("Error querying comments: %v", err)
        return nil, err
    }
    defer rows.Close()
    var comments []models.Comment
    for rows.Next() {
        var comment models.Comment
        if err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.Username, &comment.LikesCount, &comment.DislikesCount); err != nil {
            log.Printf("Error scanning comment: %v", err)
            return nil, err
        }
        comments = append(comments, comment)
    }
    if err := rows.Err(); err != nil {
        log.Printf("Error with rows: %v", err)
        return nil, err
    }
    return comments, nil
}
func SubmitComment(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.Error405(w, r)
		return
	}
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing form data:", err)
		errors.Error400(w, r)
		return
	}
	// Retrieve the comment and post ID from the form
	comment := strings.TrimSpace(r.FormValue("comment")) // Trim leading and trailing spaces
	postID := r.FormValue("post_id")
	// Input validation
	var errorsArr []string
	if comment == "" {
		errorsArr = append(errorsArr, "Comment cannot be empty or start with spaces.")
	}
	if postID == "" {
		errorsArr = append(errorsArr, "Post ID is required.")
	}
	// If there are errors, handle them
	if len(errorsArr) > 0 {
		log.Println("Validation errors:", errorsArr)
		http.Error(w, strings.Join(errorsArr, ", "), http.StatusBadRequest)
		return
	}
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Println("Error retrieving session token:", err)
		errors.Error401(w, r)
		return
	}
	// Retrieve the user ID from the sessionStore
	userID, ok := sessionStore[cookie.Value]
	if !ok {
		log.Println("Invalid session token.")
		errors.Error401(w, r)
		return
	}
	// Insert the comment into the database
	_, err = db.Exec("INSERT INTO comments (user_id, post_id, content) VALUES (?, ?, ?)", userID, postID, comment)
	if err != nil {
		log.Println("Error inserting comment into database:", err)
		errors.Error500(w, r)
		return
	}
	// Redirect to the post page
	http.Redirect(w, r, fmt.Sprintf("/postpage?id=%s", postID), http.StatusSeeOther)
}
func convertToLocalTime(t time.Time) time.Time { //UTC = +3
	loc, err := time.LoadLocation("Asia/Baghdad")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return t // return original time if there's an error
	}
	return t.In(loc) // Convert time to UTC+3
}