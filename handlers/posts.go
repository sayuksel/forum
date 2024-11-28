package handlers
import (
	"database/sql"
	"forum/errors"
	"forum/models"
	"html/template"
	"log"
	"net/http"
	"strings"
)
var funcMap = template.FuncMap{
	"split": strings.Split,
}
var templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
// Home handles the home page and displays posts with optional filtering.
func Home(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	var user models.User
	isAuthenticated := false
	userID := 0
	if err == nil {
		// Validate session
		userIDVal, ok := sessionStore[cookie.Value]
		if ok {
			isAuthenticated = true
			userID = int(userIDVal)
			err := db.QueryRow("SELECT id, username FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username)
			if err != nil {
				log.Println("Error retrieving user data:", err)
			}
		}
	}
	// Retrieve the filter parameters from the query string
	filterType := r.URL.Query().Get("filter")
	filterValue := r.URL.Query().Get("value")
	// Fetch posts based on the filter
	posts, err := fetchPosts(filterType, filterValue, userID)
	if err != nil {
		log.Println("Error fetching posts:", err)
		errors.Error500(w, r)
		return
	}
	// Prepare the data for the template
	data := models.HomePageData{
		User:            user,
		IsAuthenticated: isAuthenticated,
		Posts:           posts,
	}
	// Render the home page template
	if err := templates.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Println("Error rendering home page template:", err)
		errors.Error500(w, r)
		return
	}
}
// Post handles the post creation page.
func Post(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "post.html", nil); err != nil {
		log.Println("Error rendering post creation template:", err)
		errors.Error500(w, r)
	}
}
// SubmitPost handles the creation of a new post with categories.
func SubmitPost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.Error405(w, r)
		return
	}
	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing form data:", err)
		errors.Error400(w, r)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories[]"]
	// Trim leading/trailing spaces
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	// Validate inputs
	if title == "" || content == "" {
		log.Println("Validation error: Title or content cannot be empty.")
		http.Error(w, "Title or content cannot be empty.", http.StatusBadRequest)
		return
	}
	if len(categories) == 0 {
		log.Println("Validation error: At least one category must be selected.")
		http.Error(w, "At least one category must be selected.", http.StatusBadRequest)
		return
	}
	// Get session token
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Println("Error retrieving session token:", err)
		errors.Error401(w, r)
		return
	}
	// Validate session
	userID, ok := sessionStore[cookie.Value]
	if !ok {
		log.Println("Invalid session token.")
		errors.Error401(w, r)
		return
	}
	// Insert post into the database
	res, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
	if err != nil {
		log.Println("Error inserting post into database:", err)
		errors.Error500(w, r)
		return
	}
	// Get the inserted post ID
	postID64, err := res.LastInsertId()
	if err != nil {
		log.Println("Error retrieving last insert ID:", err)
		errors.Error500(w, r)
		return
	}
	postID := int(postID64)
	// Insert categories into the post_categories table
	for _, category := range categories {
		var categoryID int
		err := db.QueryRow("SELECT id FROM categories WHERE name = ?", category).Scan(&categoryID)
		if err == sql.ErrNoRows {
			// Insert the category if it doesn't exist
			res, err := db.Exec("INSERT INTO categories (name) VALUES (?)", category)
			if err != nil {
				log.Println("Error inserting new category:", err)
				errors.Error500(w, r)
				return
			}
			categoryID64, _ := res.LastInsertId()
			categoryID = int(categoryID64)
		} else if err != nil {
			log.Println("Error querying category ID:", err)
			errors.Error500(w, r)
			return
		}
		// Link the post with the category
		_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
		if err != nil {
			log.Println("Error inserting into post_categories table:", err)
			errors.Error500(w, r)
			return
		}
	}
	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
// fetchPosts fetches posts with their associated categories.
func fetchPosts(filterType string, filterValue string, userID int) ([]models.Posts, error) {
	var rows *sql.Rows
	var err error
	switch filterType {
	case "category": // Filter by category
		rows, err = db.Query(`
			SELECT p.id, p.user_id, u.username, p.title, p.content, IFNULL(GROUP_CONCAT(c.name), '') AS categories, p.created_at, p.likes_count, p.dislikes_count
			FROM posts p
			JOIN users u ON p.user_id = u.id
			JOIN post_categories pc ON p.id = pc.post_id
			JOIN categories c ON pc.category_id = c.id
			WHERE c.name = ?
			GROUP BY p.id
			ORDER BY p.created_at DESC`, filterValue)
	case "created_posts": // Filter by posts created by the logged-in user
		rows, err = db.Query(`
			SELECT p.id, p.user_id, u.username, p.title, p.content, IFNULL(GROUP_CONCAT(c.name), '') AS categories, p.created_at, p.likes_count, p.dislikes_count
			FROM posts p
			JOIN users u ON p.user_id = u.id
			LEFT JOIN post_categories pc ON p.id = pc.post_id
			LEFT JOIN categories c ON pc.category_id = c.id
			WHERE p.user_id = ?
			GROUP BY p.id
			ORDER BY p.created_at DESC`, userID)
	case "liked_posts": // Filter by posts liked by the logged-in user
		rows, err = db.Query(`
			SELECT p.id, p.user_id, u.username, p.title, p.content, IFNULL(GROUP_CONCAT(c.name), '') AS categories, p.created_at, p.likes_count, p.dislikes_count
			FROM posts p
			JOIN users u ON p.user_id = u.id
			LEFT JOIN post_categories pc ON p.id = pc.post_id
			LEFT JOIN categories c ON pc.category_id = c.id
			JOIN post_likes pl ON pl.post_id = p.id
			WHERE pl.user_id = ? AND pl.like = TRUE
			GROUP BY p.id
			ORDER BY p.created_at DESC`, userID)
	default: // Fetch all posts
		rows, err = db.Query(`
			SELECT p.id, p.user_id, u.username, p.title, p.content, IFNULL(GROUP_CONCAT(c.name), '') AS categories, p.created_at, p.likes_count, p.dislikes_count
			FROM posts p
			JOIN users u ON p.user_id = u.id
			LEFT JOIN post_categories pc ON p.id = pc.post_id
			LEFT JOIN categories c ON pc.category_id = c.id
			GROUP BY p.id
			ORDER BY p.created_at DESC`)
	}
	if err != nil {
		log.Println("Error executing query in fetchPosts:", err)
		return nil, err
	}
	defer rows.Close()
	var posts []models.Posts
	for rows.Next() {
		var post models.Posts
		if err := rows.Scan(&post.ID, &post.User_ID, &post.Username, &post.Title, &post.Content, &post.Category, &post.CreatedAt, &post.LikesCount, &post.DislikesCount); err != nil {
			log.Println("Error scanning row in fetchPosts:", err)
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
