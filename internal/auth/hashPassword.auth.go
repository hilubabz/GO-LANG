package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error){
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil{
		return "", err
	}
	return hash, nil
}