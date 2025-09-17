package jwt

type InDTO struct {
	ID    int    `json:"user_id"`
	Login string `json:"user_login"`
}
