package payload

type TokenRequest struct {
	AccountName string `json:"account_name"`
	SecretKey   string `json:"secret_key"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
