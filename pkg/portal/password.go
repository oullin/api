package portal

import "golang.org/x/crypto/bcrypt"

type Password struct {
	hash []byte
	seed string
}

func MakePassword(seed string) (Password, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(seed), 10)

	if err != nil {
		return Password{}, err
	}

	return Password{
		hash: bytes,
		seed: seed,
	}, nil
}

func (receiver Password) Is(seed string) bool {
	err := bcrypt.CompareHashAndPassword(receiver.hash, []byte(seed))

	return err == nil
}

func (receiver Password) GetHash() string {
	return string(receiver.hash)
}
