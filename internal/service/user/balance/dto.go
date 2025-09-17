package balance

type GetUserBalanceDTO struct {
	TotalWithdrawn int     `json:"withdrawn"`
	CurrentBalance float32 `json:"current"`
}
