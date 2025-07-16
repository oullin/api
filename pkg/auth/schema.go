package auth

const (
	PublicKeyPrefix = "pk_"
	SecretKeyPrefix = "sk_"
	LevelPublic     = "public"
	LevelSecret     = "secret"
	TokenMinLength  = 16
)

type Token struct {
	AccountName string `validate:"required,min=5"`
	PublicKey   string `validate:"required,len=16"`
	PrivateKey  string `validate:"required,len=16"`
	Length      int    `validate:"required"`
}

type ValidatedToken struct {
	AuthLevel string `validate:"required,oneof=public secret"`
	Token     string `validate:"required,len=16"`
}
