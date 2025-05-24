package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"ewallet-backend/db"

	"github.com/gorilla/mux"
)

type UserData struct {
	ID        string       `json:"id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	PublicKey string    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryName := r.URL.Query().Get("name")
	var rows *sql.Rows
	var err error

	if queryName != "" {
		rows, err = db.DB.Query(`
			SELECT id, name, balance, public_key, created_at
			FROM users
			WHERE name ILIKE $1
		`, "%"+queryName+"%")
	} else {
		rows, err = db.DB.Query(`
			SELECT id, name, balance, public_key, created_at
			FROM users
		`)
	}

	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserData
	for rows.Next() {
		var user UserData
		if err := rows.Scan(&user.ID, &user.Name, &user.Balance, &user.PublicKey, &user.CreatedAt); err != nil {
			http.Error(w, "Error scanning user " + err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func SingleUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user UserData
	err := db.DB.QueryRow(`
		SELECT id, name, balance, public_key, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Balance, &user.PublicKey, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
