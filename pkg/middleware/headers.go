package middleware

type AuthTokenHeaders struct {
	AccountName string
	PublicKey   string
	Signature   string
	Timestamp   string
	Nonce       string
	ClientIP    string
	RequestID   string
}
