package main

import (
	"context"
	"log"
),
	"context",
	"github.com/jackc/pgx/v5/pgxpool"
)
var DB *pgxpool.Pool
func connect(string url){
	var err error
	DB,err=pgx.pool.New(context.Background(),url)
	if err !=nil{
		log.Fatal(err)
	}
	if err=DB.PIng(context.Background());err!=nil{
		log.Fatal(err)
	}
	log.Println("Connected to PostgreSQL")
}

