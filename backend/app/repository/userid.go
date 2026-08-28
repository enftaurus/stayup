package repository

import (
	"context"

	"github.com/enftaurus/stayup/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func UserID(githubID int64, ctx context.Context) (pgtype.UUID, error) {
	var userID pgtype.UUID

	err := utils.DB.QueryRow(
		ctx,
		`SELECT id FROM users WHERE github_id = $1`,
		githubID,
	).Scan(&userID)

	if err != nil {
		return pgtype.UUID{}, err
	}

	return userID, nil
}
