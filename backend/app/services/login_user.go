package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/enftaurus/stayup/app/models"
	"github.com/enftaurus/stayup/app/repository"
	"github.com/enftaurus/stayup/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func hash_token(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
func user_login(uuid pgtype.UUID, metadata models.Userip, ctx context.Context) (string, error) {
	token, err := utils.GenerateRefreshToken()
	if err != nil {
		log.Println("unable to generate a reftesh token", err)
		return "", err
	}
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().UTC().Add(5 * 24 * time.Hour),
		Valid: true,
	}
	hashedToken := hash_token(token)
	user_session := models.User_session{
		UserId:       uuid,
		RefreshToken: &hashedToken,
		ExpiresAt:    &expiresAt,
		Ip:           &metadata.Ip,
		UserAgent:    metadata.UserAgent,
	}
	err := repository.UserSession(user_session, ctx)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return token, nil
}
