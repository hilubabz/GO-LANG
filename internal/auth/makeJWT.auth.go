package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error){
	now := time.Now().UTC()
	if expiresIn==time.Duration(0){
		expiresIn=time.Duration(60*60)*time.Second
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:"chirpy-access",	
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject: userID.String(),
	})
	token, err := t.SignedString([]byte(tokenSecret))
	if err!=nil{
		return "", err
	}
	return token, nil
}