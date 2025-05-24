package model

type TransferRequest struct {
	ToUserID int     `json:"to_user_id"`
	Amount   float64 `json:"amount"`
}
