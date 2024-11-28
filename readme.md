# Forum Application

This project is a simple forum application with users, posts, comments, and likes/dislikes functionality. Below are the steps to set up the database and run the application.

## Prerequisites

Before you begin, ensure you have the following installed on your machine:

- [SQLite3](https://www.sqlite.org/download.html)
- Go (version 1.16 or higher) - [Download Go](https://golang.org/dl/)

## Database Setup

1. **Create the Database**:
   Open your terminal and navigate to the project directory. Run the following command to create the database file:

   ```bash
   sqlite3 forum.db
   ```

2. **Create Tables**:
   Once in the SQLite shell, execute the following SQL commands to create the necessary tables:

   ```sql
   -- Users table
   CREATE TABLE IF NOT EXISTS users (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       username TEXT NOT NULL UNIQUE,
       email TEXT NOT NULL UNIQUE,
       password TEXT NOT NULL,
       session_token TEXT,
       session_expiry DATETIME
   );
   ```

3. **Exit SQLite**:
   After executing the SQL commands, type `.exit` to exit the SQLite shell.

## Running the Application

1. **Run the Go Application**:
   In your terminal, make sure you are in the project directory (where `main.go` is located), and then execute the following command to run the application:

   ```bash
   go run main.go
   ```