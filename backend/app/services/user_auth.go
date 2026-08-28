package services

import (
	"context"
	"errors"
	"log"

	"github.com/enftaurus/stayup/app/models"
	"github.com/enftaurus/stayup/app/repository"
	"github.com/jackc/pgx/v5"
)

func UserAuth(GithubUser models.GithubUser, metadata models.Userip, ctx context.Context) (string, error) {
	userid, err := repository.UserID(GithubUser.GithubID, ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			userid, err = register_user(GithubUser, ctx)
			if err != nil {
				log.Println("unable to register", err)
				return "", err
			}
		}
		return "", err
	}

	hashed_token, err := user_login(userid, metadata, ctx)
	if err != nil {
		log.Println("unable to generate refresh token :", err)
		return "", err
	}
	return hashed_token, nil

}
