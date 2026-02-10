package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/config"
	"github.com/m4xvel/monetych_bot/internal/crypto"
	"github.com/m4xvel/monetych_bot/internal/delivery/telegram"
	"github.com/m4xvel/monetych_bot/internal/infra"
	"github.com/m4xvel/monetych_bot/internal/logger"
	"github.com/m4xvel/monetych_bot/internal/repository/postgres"
	"github.com/m4xvel/monetych_bot/internal/usecase"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Fatalw("panic occurred", "panic", r)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger.Init(logger.Config{
		Env: cfg.Env, // dev | prod
	})
	defer logger.Sync()

	logger.Log.Infow("application starting",
		"env", cfg.Env,
		"debug", cfg.Debug,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := infra.NewPostgresPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatalw("failed to connect to database", "err", err)
	}
	defer pool.Close()

	logger.Log.Infow("database connected")

	bot, err := telegram.NewBot(cfg.BotToken, cfg.Debug)
	if err != nil {
		logger.Log.Fatalw("failed to initialize telegram bot", "err", err)
	}

	logger.Log.Infow("telegram bot initialized",
		"bot_username", bot.Self.UserName,
	)

	keyBase64, err := crypto.New(cfg.KeyBase64)
	if err != nil {
		logger.Log.Fatalw("failed to initialize key crypto", "err", err)
	}

	userRepo := postgres.NewUserRepo(pool)
	stateRepo := postgres.NewUserStateRepo(pool)
	gameRepo := postgres.NewGameRepo(pool)
	orderRepo := postgres.NewOrderRepo(pool, keyBase64)
	expertRepo := postgres.NewExpertRepo(pool)
	supportRepo := postgres.NewSupportRepo(pool)
	orderMessageRepo := postgres.NewOrderMessageRepo(pool)
	orderChatMessageRepo := postgres.NewOrderChatMessagesRepo(pool, keyBase64)
	reviewRepo := postgres.NewReviewRepo(pool)
	callbackTokenRepo := postgres.NewCallbackTokenRepo(pool)
	userPolicyAcceptancesRepo := postgres.NewUserPolicyAcceptancesRepo(pool)

	userService := usecase.NewUserService(userRepo)
	stateService := usecase.NewStateService(stateRepo)
	gameService := usecase.NewGameService(gameRepo)
	orderService := usecase.NewOrderService(orderRepo, userRepo)
	expertService := usecase.NewExpertService(expertRepo)
	supportService := usecase.NewSupportService(supportRepo)
	orderMessageService := usecase.NewOrderMessageService(orderMessageRepo)
	orderChatMessageService := usecase.
		NewOrderChatMessageService(orderChatMessageRepo)
	reviewService := usecase.NewReviewService(reviewRepo)
	callbackTokenService := usecase.NewCallbackTokenService(callbackTokenRepo)
	userPolicyAcceptancesService := usecase.
		NewUserPolicyAcceptancesService(userPolicyAcceptancesRepo)

	if err := gameService.InitCache(ctx); err != nil {
		logger.Log.Fatalw("failed to init game cache", "err", err)
	}

	if err := expertService.InitCache(ctx); err != nil {
		logger.Log.Fatalw("failed to init expert cache", "err", err)
	}

	if err := supportService.InitCache(ctx); err != nil {
		logger.Log.Fatalw("failed to init support cache", "err", err)
	}

	logger.Log.Infow("caches initialized")

	handler := telegram.NewHandler(
		bot,
		userService,
		stateService,
		gameService,
		orderService,
		expertService,
		supportService,
		reviewService,
		orderMessageService,
		orderChatMessageService,
		callbackTokenService,
		userPolicyAcceptancesService,
		cfg.PrivacyPolicyURL,
		cfg.PublicOfferURL,
	)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	logger.Log.Infow("bot started, listening for updates")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Infow("shutdown requested")
			return
		case update, ok := <-updates:
			if !ok {
				logger.Log.Warnw("updates channel closed")
				return
			}
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Log.Errorw("panic in update handler", "panic", r)
					}
				}()

				handler.Route(ctx, update)
			}()
		}
	}
}
