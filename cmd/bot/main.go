package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
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

	go runOrderMessagesCleanup(
		ctx,
		orderMessageService,
		cfg.OrderMsgRetentionDays,
	)

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
		cfg.VerificationEnabled,
		cfg.PrivacyPolicyURL,
		cfg.PublicOfferURL,
	)

	updates, stopUpdates, err := setupUpdatesSource(ctx, bot, cfg)
	if err != nil {
		logger.Log.Fatalw("failed to configure updates source", "err", err)
	}
	defer stopUpdates()

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

func setupUpdatesSource(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	cfg *config.Config,
) (tgbotapi.UpdatesChannel, func(), error) {
	if !cfg.WebhookEnabled {
		_, err := bot.Request(tgbotapi.DeleteWebhookConfig{
			DropPendingUpdates: false,
		})
		if err != nil {
			logger.Log.Warnw("failed to delete webhook before polling",
				"err", err,
			)
		}

		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		logger.Log.Infow("updates source configured",
			"mode", "polling",
		)

		return bot.GetUpdatesChan(u), bot.StopReceivingUpdates, nil
	}

	webhookConfig, err := tgbotapi.NewWebhook(cfg.WebhookURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse TELEGRAM_WEBHOOK_URL: %w", err)
	}

	if webhookConfig.URL.Scheme != "https" {
		return nil, nil, fmt.Errorf("TELEGRAM_WEBHOOK_URL must use https, got %q", webhookConfig.URL.Scheme)
	}

	if webhookConfig.URL.Path == "" || webhookConfig.URL.Path == "/" {
		return nil, nil, errors.New("TELEGRAM_WEBHOOK_URL must include a non-root path")
	}

	webhookConfig.DropPendingUpdates = cfg.WebhookDropPending

	if _, err := bot.Request(webhookConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to set telegram webhook: %w", err)
	}

	updates := bot.ListenForWebhook(webhookConfig.URL.Path)

	listener, err := net.Listen("tcp", cfg.WebhookListenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen for webhook on %s: %w", cfg.WebhookListenAddr, err)
	}

	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Log.Errorw("failed to shutdown webhook server",
					"err", err,
				)
			}
		case err := <-serverErr:
			if err != nil {
				logger.Log.Errorw("webhook server stopped unexpectedly",
					"err", err,
				)
			}
		}
	}()

	logger.Log.Infow("updates source configured",
		"mode", "webhook",
		"webhook_url", cfg.WebhookURL,
		"listen_addr", cfg.WebhookListenAddr,
		"path", webhookConfig.URL.Path,
	)

	stop := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Errorw("failed to stop webhook server",
				"err", err,
			)
		}
	}

	return updates, stop, nil
}

func runOrderMessagesCleanup(
	ctx context.Context,
	service *usecase.OrderMessageService,
	retentionDays int,
) {
	if retentionDays <= 0 {
		logger.Log.Infow("order messages cleanup disabled",
			"retention_days", retentionDays,
		)
		return
	}

	run := func() {
		purgeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		before := time.Now().AddDate(0, 0, -retentionDays)
		deletedRows, err := service.PurgeDeletedBefore(purgeCtx, before)
		if err != nil {
			logger.Log.Errorw("order messages cleanup failed",
				"retention_days", retentionDays,
				"before", before,
				"err", err,
			)
			return
		}

		if deletedRows > 0 {
			logger.Log.Infow("order messages cleaned",
				"deleted_rows", deletedRows,
				"retention_days", retentionDays,
			)
		}
	}

	run()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
