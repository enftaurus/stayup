package main

import (
	"log"

	"github.com/enftaurus/stayup/config"
	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	err := utils.Connect(config.GetEnv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer utils.DB.Close()
	log.Println("Connected to PostgreSQL")
	router := gin.Default()
	RegisterRoutes(router)
	router.Run(":8080")
}
