package views

type UserBalance struct {
	UserID         int     `json:"user_id"`
	TotalAccrued   int     `json:"total_accrued"`
	TotalWithdrawn int     `json:"total_withdrawn"`
	CurrentBalance float32 `json:"current_balance"`
}
