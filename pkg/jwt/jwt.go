package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(id int64, username, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"id":       id,
		"username": username,
		"exp":      jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key := []byte(secretKey)

	tokenStr, err := token.SignedString(key)

	return tokenStr, err
}

func ValidateToken(tokenStr, secretKey string, withClaimValidation bool) (int64, string, error) {
	var (
		key    = []byte(secretKey)
		claims = jwt.MapClaims{}
		token  *jwt.Token
		err    error
	)

	if withClaimValidation {
		token, err = jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return key, nil
		})
	} else {
		token, err = jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return key, nil
		}, jwt.WithoutClaimsValidation())
	}

	if err != nil {
		return 0, "", err
	}

	if !token.Valid {
		return 0, "", errors.New("invalid token")
	}

	// userID := int64(claims["id"].(float64))
	// username := claims["username"].(string)

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, "", errors.New("invalid id claim")
	}
	userID := int64(idFloat)

	username, ok := claims["username"].(string)
	if !ok {
		return 0, "", errors.New("invalid username claim")
	}

	return userID, username, nil
}
