package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID             pgtype.UUID        `db:"id" json:"id"`
	GithubID       int64              `db:"github_id" json:"github_id,omitempty"`
	GithubUsername string             `db:"github_username" json:"github_username,omitempty"`
	GithubName     *string            `db:"github_name" json:"github_name,omitempty"`
	Email          *string            `db:"email" json:"email,omitempty"`
	AvatarURL      *string            `db:"avatar_url" json:"avatar_url,omitempty"`
	CreatedAt      time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `db:"updated_at" json:"updated_at"`
	LastLoginAt    pgtype.Timestamptz `db:"last_login_at" json:"last_login_at"`
}
type GithubUser struct {
	GithubID       int64   `db:"github_id" json:"id"`
	GithubUsername string  `db:"github_username" json:"login"`
	GithubName     string  `db:"github_name" json:"name"`
	Email          *string `db:"email" json:"email"`
	AvatarURL      *string `db:"avatar_url" json:"avatar_url"`
}
