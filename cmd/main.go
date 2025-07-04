package main

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wbtech"
	"wbtech/internal/cache"
	"wbtech/internal/handler"
	"wbtech/internal/kafka"
	"wbtech/internal/repository"
	"wbtech/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := initConfig(); err != nil {
		slog.Error("init config err: %v", err)
		os.Exit(1)
	}

	if err := godotenv.Load(".env"); err != nil {
		slog.Error("load .env err: %v", err)
		os.Exit(1)
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		slog.Error("init db err: %v", err)
		os.Exit(1)
	}

	orderCache := cache.NewOrderCache(
		viper.GetInt("cache.limit"),
		viper.GetString("cache.dump"),
		viper.GetDuration("cache.ttl"),
	)
	defer func() {
		if err := orderCache.Close(); err != nil {
			slog.Error("failed to save cache dump", "error", err)
		}
	}()

	repos := repository.NewRepository(db)
	services := service.NewService(repos, orderCache)
	handlers := handler.NewHandler(services)

	kafkaConsumer := kafka.NewConsumer(
		viper.GetStringSlice("kafka.brokers"),
		viper.GetString("kafka.topic"),
		viper.GetString("kafka.group_id"),
		services.Order.(*service.OrderService),
	)
	defer func(kafkaConsumer *kafka.Consumer) {
		err := kafkaConsumer.Close()
		if err != nil {
			slog.Error("failed to close kafka consumer", "error", err)
		}
	}(kafkaConsumer)

	kafkaCtx, kafkaCancel := context.WithCancel(context.Background())
	defer kafkaCancel()
	go kafkaConsumer.Consume(kafkaCtx)

	srv := new(wbtech.Server)
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Graceful shutdown:
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("Shutting down gracefully...")
	case err := <-serverErr:
		slog.Error("Server runtime error", "error", err)
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error:", "error", err)
	}

	slog.Info("Server shutdown completed")
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
