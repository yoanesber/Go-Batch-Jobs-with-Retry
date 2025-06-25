package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yoanesber/go-batch-jobs-with-retry/config/database"
	"github.com/yoanesber/go-batch-jobs-with-retry/internal/repository"
	"github.com/yoanesber/go-batch-jobs-with-retry/internal/service"
	"github.com/yoanesber/go-batch-jobs-with-retry/pkg/logger"
	"github.com/yoanesber/go-batch-jobs-with-retry/pkg/scheduler"
)

var (
	dbInitialized     bool
	loggerInitialized bool
)

func main() {
	// Create base context with cancel for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init all dependencies
	initializeDependencies()

	// Initialize the scheduler
	initializeScheduler()

	// Graceful shutdown
	gracefulShutdown(cancel)

	// Get environment variables
	env := os.Getenv("ENV")
	port := os.Getenv("PORT")
	isSSL := os.Getenv("IS_SSL")
	apiVersion := os.Getenv("API_VERSION")

	if env == "" || port == "" || isSSL == "" || apiVersion == "" {
		logger.Panic("One or more required environment variables are not set", logrus.Fields{
			"env":        env,
			"port":       port,
			"isSSL":      isSSL,
			"apiVersion": apiVersion,
		})

		return
	}

	// Set Gin mode
	gin.SetMode(gin.DebugMode)
	if env == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	r := gin.Default()
	r.SetTrustedProxies(nil) // Set trusted proxies to nil to avoid issues with forwarded headers

	// Start the server
	if err := r.Run(":" + port); err != nil {
		logger.Panic("Failed to start server", logrus.Fields{
			"port":  port,
			"error": err,
		})

		return
	}
}

func initializeDependencies() {
	if !loggerInitialized {
		if !logger.Init() {
			logger.Panic("Failed to initialize logger", nil)
			return
		} else {
			loggerInitialized = true
		}
	}
	if !dbInitialized {
		if !database.InitPostgres() {
			logger.Panic("Failed to initialize Postgres database", nil)
			return
		} else {
			dbInitialized = true
		}
	}
}

func initializeScheduler() {
	// add a job to the scheduler
	trxRepo := repository.NewTransactionRepository()
	jobRepo := repository.NewBatchJobExecutionRepository()
	jobService := service.NewBatchJobExecutionService(trxRepo, jobRepo)
	scheduler.DurationJob(0, 0, 120, jobService.ProcessLargeBatch, "transaction")
	// scheduler.DailyJob(20, 26, 0, jobService.ProcessLargeBatch, "transaction")
	// scheduler.WeeklyJob(1, time.Tuesday, 20, 48, 0, jobService.ProcessLargeBatch, "transaction")
	// scheduler.MonthlyJob(1, 24, 21, 0, 0, jobService.ProcessLargeBatch, "transaction")
}

func gracefulShutdown(cancel context.CancelFunc) {
	// Handle graceful shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		logger.Info("Received signal for graceful shutdown", logrus.Fields{
			"signal": sig,
		})

		// Cancel context
		cancel()

		// Clean up resources
		if dbInitialized {
			logger.Info("Closing database connection", nil)
			database.ClosePostgres()
		}

		logger.Info("Server shutting down gracefully", nil)
		os.Exit(0)
	}()
}
