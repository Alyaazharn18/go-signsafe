package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"ewallet-backend/db"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"public_key"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "cannot start transaction", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.PublicKey == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var existingUserID string // Untuk menampung ID jika user ditemukan
	err = tx.QueryRow(`SELECT id FROM users WHERE name = $1`, req.Name).Scan(&existingUserID)

	if err == nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Username '%s' is already taken. Please choose another one.", req.Name), http.StatusConflict)
		log.Printf("Registration denied: Username '%s' already exists (ID: %s)", req.Name, existingUserID)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		http.Error(w, "Internal server error during username check.", http.StatusInternalServerError)
		log.Printf("Database error while checking username existence: %v", err)
		return
	}

	newUUID := uuid.New().String()
	currentTime := time.Now()
	initialBalance := 0.00

	_, err = tx.Exec(`
			INSERT INTO users (id, name, public_key, balance, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, newUUID, req.Name, req.PublicKey, initialBalance, currentTime)
	if err != nil {
		http.Error(w, "User already exists or database error", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
		return
	}

	userResponse := UserResponse{
		ID:        newUUID,
		Name:      req.Name,
		PublicKey: req.PublicKey,
		Balance:   initialBalance,
		CreatedAt: currentTime,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if encodeErr := json.NewEncoder(w).Encode(userResponse); encodeErr != nil {
		log.Printf("Error encoding user response for user '%s': %v", newUUID, encodeErr)
	}

	log.Printf("User '%s' registered successfully with ID: %s. Response sent.", req.Name, newUUID)
}
