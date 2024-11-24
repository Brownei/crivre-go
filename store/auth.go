package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brownei/crivre-go/utils"
)

type AuthStore struct {
	db *sql.DB
}

func (s *AuthStore) Login(ctx context.Context, dbPassword, payloadPassword, email string) (string, error) {
	err := utils.VerifyPassword(dbPassword, payloadPassword)
	if err != nil {
		return "", fmt.Errorf("Password Incorrect")
	}

	token := utils.JwtToken(email, ctx)

	return token, nil
}
