package auth

const (
	PublicKeyPrefix      = "pk_"
	SecretKeyPrefix      = "sk_"
	TokenMinLength       = 16
	AccountNameMinLength = 5
	EncryptionKeyLength  = 32
)

type Token struct {
	AccountName        string
	KeyID              string
	PublicKey          string
	EncryptedPublicKey []byte
	SecretKey          string
	EncryptedSecretKey []byte
}

type SecureToken struct {
	PlainText     string
	EncryptedText []byte
}
