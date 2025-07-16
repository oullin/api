package auth

const (
	PublicKeyPrefix      = "pk_"
	SecretKeyPrefix      = "sk_"
	TokenMinLength       = 16
	AccountNameMinLength = 5
)

type Token struct {
	AccountName string `validate:"required,min=5"`
	PublicKey   string `validate:"required"`
	SecretKey   string `validate:"required"`
	Length      int    `validate:"required"`
}
