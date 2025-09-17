package order

import "time"

type orderDTO struct {
	OrderNumber string    `json:"number"`
	Status      string    `json:"status"`
	Accrual     float32   `json:"accrual"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
