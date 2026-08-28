package services

import (
	"context"
	"log"

	"github.com/enftaurus/stayup/app/models"
	"github.com/enftaurus/stayup/app/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func register_user(user models.GithubUser, ctx context.Context) (pgtype.UUID, error) {

	uuid, err := repository.InsertUser(ctx, user)
	if err != nil {
		log.Println("unable to insert user")
		return pgtype.UUID{}, err
	}
	return uuid, nil
}
