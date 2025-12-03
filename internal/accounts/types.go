package accounts

type Account struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AuthType  string `json:"auth_type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
