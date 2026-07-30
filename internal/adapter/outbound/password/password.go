package password

import (
	"golang.org/x/crypto/bcrypt"
)

type hasher struct{}

func NewHasher() *hasher { return &hasher{} }

func (h *hasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (h *hasher) Compare(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
