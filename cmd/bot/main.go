package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/config"
	"github.com/m4xvel/monetych_bot/internal/delivery/telegram"
	"github.com/m4xvel/monetych_bot/internal/infra"
	"github.com/m4xvel/monetych_bot/internal/repository/postgres"
	"github.com/m4xvel/monetych_bot/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := infra.NewPostgresPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	bot, err := telegram.NewBot(cfg.BotToken, cfg.Debug)
	if err != nil {
		log.Fatalf("failed to initialize bot: %v", err)
	}

	userRepo := postgres.NewUserRepo(pool)
	stateRepo := postgres.NewUserStateRepo(pool)
	gameRepo := postgres.NewGameRepo(pool)
	orderRepo := postgres.NewOrderRepo(pool)
	expertRepo := postgres.NewExpertRepo(pool)
	orderMessageRepo := postgres.NewOrderMessageRepo(pool)
	reviewRepo := postgres.NewReviewRepo(pool)

	userService := usecase.NewUserService(userRepo)
	stateService := usecase.NewStateService(stateRepo)
	gameService := usecase.NewGameService(gameRepo)
	orderService := usecase.NewOrderService(orderRepo, userRepo)
	expertService := usecase.NewExpertService(expertRepo)
	orderMessageService := usecase.NewOrderMessageService(orderMessageRepo)
	reviewService := usecase.NewReviewService(reviewRepo)

	if err := gameService.InitCache(ctx); err != nil {
		log.Fatalf("failed to init game cache: %v", err)
	}

	if err := expertService.InitCache(ctx); err != nil {
		log.Fatalf("failed to init expert cache: %v", err)
	}

	handler := telegram.NewHandler(
		bot,
		userService,
		stateService,
		gameService,
		orderService,
		expertService,
		reviewService,
		orderMessageService,
	)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("shutdown requested, exiting")
			return
		case update, ok := <-updates:
			if !ok {
				log.Println("updates channel closed")
				return
			}
			go handler.Route(ctx, update)
		}
	}
}
