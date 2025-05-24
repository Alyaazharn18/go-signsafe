package model

type TransferRequest struct {
	ToUserID string  `json:"to_user_id"`
	Amount   float64 `json:"amount"`
}
