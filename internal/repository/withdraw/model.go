package withdraw

import "time"

type NewWithdraw struct {
	UserId      string  `json:"user_id"`
	OrderNumber string  `json:"order_number"`
	Sum         float32 `json:"sum"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
