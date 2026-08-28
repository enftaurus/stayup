package utils

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(url string) error {
	var err error
	DB, err = pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}
	if err = DB.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to PostgreSQL")
	return nil
}
