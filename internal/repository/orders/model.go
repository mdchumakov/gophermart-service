package orders

import "time"

type Order struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
	Accrual     float32   `json:"accrual"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
