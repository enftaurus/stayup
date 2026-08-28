package repository

import (
	"context"

	"github.com/enftaurus/stayup/app/models"
	"github.com/enftaurus/stayup/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func InsertUser(ctx context.Context, user models.GithubUser) (pgtype.UUID, error) {
	var userID pgtype.UUID

	err := utils.DB.QueryRow(
		ctx,
		`INSERT INTO users (
			github_id,
			github_username,
			github_name,
			email,
			avatar_url
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		user.GithubID,
		user.GithubUsername,
		user.GithubName,
		user.Email,
		user.AvatarURL,
	).Scan(&userID)

	if err != nil {
		return pgtype.UUID{}, err
	}

	return userID, nil
}
