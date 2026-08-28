package repository

import (
	"context"
	"log"

	"github.com/enftaurus/stayup/app/models"
	"github.com/enftaurus/stayup/utils"
)

func UserSession(session_details models.User_session, ctx context.Context) error {
	_, err := utils.DB.Exec(
		ctx,
		`INSERT INTO user_sessions (
            user_id,refresh_token_hash,expires_at,ip_address,user_agent ) VALUES ($1,$2,$3,$4,$5)`,
		session_details.UserId,
		session_details.RefreshToken,
		session_details.ExpiresAt,
		session_details.Ip,
		session_details.UserAgent,
	)
	if err != nil {
		log.Println("unable to insert session details into db ", err)
		return err

	}
	return nil
}
