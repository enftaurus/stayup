package models

import (
	"net"

	"github.com/jackc/pgx/v5/pgtype"
)

type Userip struct {
	Ip        net.IP  `db:"ip_address" json:"ip"`
	UserAgent *string `db:"user_agent" json:"user_agent"`
}

type User_session struct {
	Id           pgtype.UUID         `db:"id"`
	UserId       pgtype.UUID         `db:"user_id"`
	RefreshToken *string             `db:"refresh_token_hash" json:"refresh_token" `
	ExpiresAt    *pgtype.Timestamptz `db:"expires_at" json:"expires"`
	UserAgent    *string             `db:"user_agent" json:"user_agent"`
	Ip           *net.IP             `db:"ip_address" json:"ip"`
	CreatedAt    *pgtype.Timestamptz `db:"created_at" json:"created_at"`
	LastUsedAt   *pgtype.Timestamptz `db:"last_used_at" json:"last_used_at"`
}
