package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"ewallet-backend/db"
)

type HistoryEntry struct {
	ID           int     `json:"id"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	ToUserID     *string `json:"to_user_id,omitempty"`
	Timestamp    string  `json:"timestamp"`
	Nonce        string  `json:"nonce"`
	Signature    string  `json:"signature"`
	BalanceAfter float64 `json:"balance_after"`
	RecieverName *string `json:"reciever_name"`
	SenderName   *string `json:"sender_name"`
}

func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := db.DB

	userID := r.Header.Get("X-UserID")

	limitStr := r.URL.Query().Get("limit")

	limit := 20
	maxLimit := 100

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = min(parsedLimit, maxLimit)
		}
	}

	query := fmt.Sprintf(`
    SELECT 
        h.id, 
        h.type, 
        h.amount, 
        h.to_user_id, 
        h.timestamp, 
        h.nonce, 
        h.signature, 
        h.balance_after, 
        sender.name AS sender_name,
        receiver.name AS receiver_name
    FROM history h
    LEFT JOIN users sender ON h.to_user_id = sender.id
    LEFT JOIN users receiver ON h.user_id = receiver.id
    
    WHERE h.user_id = $1
    ORDER BY h.timestamp DESC
    LIMIT %d
    `, limit)

	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var history []HistoryEntry
	for rows.Next() {
		var h HistoryEntry
		err := rows.Scan(&h.ID, &h.Type, &h.Amount, &h.ToUserID, &h.Timestamp, &h.Nonce, &h.Signature, &h.BalanceAfter, &h.RecieverName, &h.SenderName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		history = append(history, h)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
