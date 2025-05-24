package handlers

import (
	"encoding/json"
	"net/http"

	"ewallet-backend/auth"
	"ewallet-backend/db"
	"ewallet-backend/model"
)

func TransferHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fromUserID := auth.GetUserID(r)

	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.ToUserID == fromUserID {
		http.Error(w, "cannot transfer to yourself", http.StatusBadRequest)
		return
	}

	nonce := r.Header.Get("X-Nonce")
	signature := r.Header.Get("X-Signature")

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "cannot start transaction", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	var fromBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", fromUserID).Scan(&fromBalance)
	if err != nil {
		http.Error(w, "sender not found", http.StatusNotFound)
		return
	}

	if fromBalance < req.Amount {
		http.Error(w, "insufficient balance", http.StatusBadGateway)
		return
	}

	var toBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", req.ToUserID).Scan(&toBalance)
	if err != nil {
		http.Error(w, "recipient not found", http.StatusBadRequest)
		return
	}

	// Update saldo pengirim
	newFromBalance := fromBalance - req.Amount
	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", newFromBalance, fromUserID)
	if err != nil {
		http.Error(w, "failed to update sender balance", http.StatusInternalServerError)
		return
	}

	// Update saldo penerima
	newToBalance := toBalance + req.Amount
	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", newToBalance, req.ToUserID)
	if err != nil {
		http.Error(w, "failed to update recipient balance", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		INSERT INTO history (user_id, to_user_id, type, amount, nonce, signature, balance_after)
		VALUES ($1, $2, 'transfer', $3, $4, $5, $6)
	`, fromUserID, req.ToUserID, req.Amount, nonce, signature, newFromBalance)
	if err != nil {
		http.Error(w, "failed to log sender history", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		INSERT INTO history (user_id, to_user_id, type, amount, nonce, signature, balance_after)
		VALUES ($1, $2, 'receive', $3, $4, $5, $6)
	`, req.ToUserID, fromUserID, req.Amount, nonce, signature, newToBalance)
	if err != nil {
		http.Error(w, "failed to log receiver history", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("transfer success"))
}
