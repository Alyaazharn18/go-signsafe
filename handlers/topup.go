package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"ewallet-backend/db"
)

type TopupRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

func TopupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req TopupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "can't start db transaction", http.StatusBadRequest)
		return
	}

	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1", req.UserID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get user balance", http.StatusInternalServerError)
		return
	}

	newBalance := currentBalance + req.Amount

	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", newBalance, req.UserID)
	if err != nil {
		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
		return
	}

	nonce := r.Header.Get("X-Nonce")
	signature := r.Header.Get("X-Signature")

	_, err = tx.Exec(`
			INSERT INTO history (user_id, type, amount, nonce, signature, balance_after)
			VALUES ($1, 'topup', $2, $3, $4, $5)
		`, req.UserID, req.Amount, nonce, signature, newBalance)
	if err != nil {
		http.Error(w, "Failed to insert history", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Transaction commit failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Topup successful. New balance: %.2f", newBalance)
}
