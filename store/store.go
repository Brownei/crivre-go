package store

import (
	"context"
	"database/sql"

	"github.com/brownei/crivre-go/types"
)

type Store struct {
	User interface {
		GetChifundsUser(email string, forLogin bool) (*types.User, error)
		GetUsersByEmail(ctx context.Context, email string, forLogn bool) (*types.User, error)
		GetAllUsers() ([]types.User, error)
		CreateNewUser(ctx context.Context, payload types.RegisterUserPayload) (*types.User, error)
		CreateChiFundsAdminUser(payload types.RegisterUserPayload) error
	}

	Auth interface {
		Login(ctx context.Context, dbPassword, payloadPassword, email string) (string, error)
	}

	Category interface {
		GetAllCategories()
		CreateNewCategory()
		GetOneCategory()
		GetAllClothingReferenceToACategory()
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Auth: &AuthStore{db},
	}
}
