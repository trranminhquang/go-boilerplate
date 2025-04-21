package models

import (
	"context"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/trranminhquang/go-boilerplate/internal/db"
	"github.com/trranminhquang/go-boilerplate/pkg/crypto"
)

type User struct {
	ID uuid.UUID `json:"id" db:"id"`

	Email            db.NullString `json:"email" db:"email"`
	EmailConfirmedAt *time.Time    `json:"email_confirmed_at,omitempty" db:"email_confirmed_at"`

	Phone            db.NullString `json:"phone" db:"phone"`
	PhoneConfirmedAt *time.Time    `json:"phone_confirmed_at,omitempty" db:"phone_confirmed_at"`

	EncryptedPassword *string `json:"-" db:"encrypted_password"`

	UserMetaData JSONMap `json:"user_metadata" db:"raw_user_meta_data"`

	// For backward compatibility only. Use EmailConfirmedAt or PhoneConfirmedAt instead.
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" db:"confirmed_at" rw:"r"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name used by pop
func (User) TableName() string {
	tableName := "users"
	return tableName
}

// NewUser initializes a new user from an email, password and user data.
func NewUser(phone, email, password string, userData map[string]interface{}) (*User, error) {
	passwordHash := ""

	if password != "" {
		pw, err := crypto.GenerateFromPassword(context.Background(), password)
		if err != nil {
			return nil, err
		}
		passwordHash = pw
	}

	if userData == nil {
		userData = make(map[string]interface{})
	}

	id := uuid.Must(uuid.NewV4())
	user := &User{
		ID:                id,
		Email:             db.NullString(strings.ToLower(email)),
		Phone:             db.NullString(phone),
		UserMetaData:      userData,
		EncryptedPassword: &passwordHash,
	}

	return user, nil
}
