package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/enftaurus/stayup/config"
	"github.com/enftaurus/stayup/engine"
	"github.com/enftaurus/stayup/migrations"
	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	dbURL := config.GetEnv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://stayup:stayup_dev_password@localhost:5432/stayup?sslmode=disable"
	}

	err := utils.Connect(dbURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer utils.DB.Close()

	// Run migrations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := migrations.Run(ctx, utils.DB); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Kafka configuration
	kafkaBrokersStr := config.GetEnv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		kafkaBrokersStr = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokersStr, ",")
	topic := config.GetEnv("KAFKA_TOPIC_CHECKS")
	if topic == "" {
		topic = "stayup.checks.due"
	}

	// Ensure Kafka topic exists with 12 partitions
	if err := engine.EnsureTopic(brokers, topic, 12, 1); err != nil {
		log.Printf("Warning: EnsureTopic error: %v", err)
	}

	kafkaWriter := engine.NewKafkaWriter(brokers, topic)
	defer kafkaWriter.Close()

	// Scheduler settings
	tickSec, _ := strconv.Atoi(config.GetEnv("SCHEDULER_TICK"))
	if tickSec <= 0 {
		tickSec = 30
	}
	lookaheadSec, _ := strconv.Atoi(config.GetEnv("SCHEDULER_LOOKAHEAD"))
	if lookaheadSec <= 0 {
		lookaheadSec = 120
	}
	scheduler := engine.NewScheduler(utils.DB, kafkaWriter, tickSec, lookaheadSec)
	go scheduler.Start(ctx)

	// Worker pool settings
	workerCount, _ := strconv.Atoi(config.GetEnv("WORKER_COUNT"))
	if workerCount <= 0 {
		workerCount = 5
	}
	allowPrivate := strings.ToLower(config.GetEnv("SSRF_ALLOW_PRIVATE")) == "true"

	var wg sync.WaitGroup
	workerPool := engine.NewWorkerPool(utils.DB, brokers, topic, "stayup-workers", workerCount, allowPrivate)
	workerPool.Start(ctx, &wg)

	// Retention cleaner
	retention := engine.NewRetentionCleaner(utils.DB)
	go retention.Start(ctx)

	// Gin HTTP router
	router := gin.Default()
	RegisterRoutes(router)

	port := config.GetEnv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("StayUp API server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down StayUp services...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("StayUp services stopped gracefully.")
}
