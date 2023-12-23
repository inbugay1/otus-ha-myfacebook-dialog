package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/inbugay1/httprouter"
	"myfacebook-dialog/internal/apiclient"
	"myfacebook-dialog/internal/apiv1/handler"
	"myfacebook-dialog/internal/apiv1/middleware"
	"myfacebook-dialog/internal/config"
	"myfacebook-dialog/internal/db"
	"myfacebook-dialog/internal/httpclient"
	"myfacebook-dialog/internal/httphandler"
	httproutermiddleware "myfacebook-dialog/internal/httprouter/middleware"
	"myfacebook-dialog/internal/httpserver"
	"myfacebook-dialog/internal/myfacebookapiclient"
	"myfacebook-dialog/internal/repository/rest"
	sqlxrepo "myfacebook-dialog/internal/repository/sqlx"
)

func main() {
	if err := run(); err != nil {
		slog.Error(fmt.Sprintf("Application error: %s", err))

		os.Exit(1)
	}
}

func run() error {
	envConfig := config.GetConfigFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(envConfig.LogLevel),
	}))
	slog.SetDefault(logger)

	appDB := db.New(db.Config{
		DriverName:         envConfig.DBDriverName,
		Host:               envConfig.DBHost,
		Port:               envConfig.DBPort,
		User:               envConfig.DBUsername,
		Password:           envConfig.DBPassword,
		DBName:             envConfig.DBName,
		SSLMode:            envConfig.DBSSLMode,
		MaxOpenConnections: envConfig.DBMaxOpenConnections,
		MigrationPath:      "./storage/migrations",
	})

	if err := appDB.Connect(context.Background()); err != nil {
		return fmt.Errorf("cannot connect to appDB: %w", err)
	}

	defer func() {
		if err := appDB.Disconnect(); err != nil {
			slog.Error(fmt.Sprintf("Failed to disconnect from app db: %s", err))
		}
	}()

	if err := appDB.Migrate(); err != nil {
		return fmt.Errorf("appDB migration failed: %w", err)
	}

	httpClient := httpclient.New(&httpclient.Config{
		InsecureSkipVerify: true,
	})

	apiClient := apiclient.New(envConfig.MyfacbookAPIBaseURL, httpClient)

	myfacebookAPIClient := myfacebookapiclient.New(apiClient)

	dialogRepository := sqlxrepo.NewDialogRepository(appDB)
	userRepository := rest.NewUserRepository(myfacebookAPIClient)

	router := httprouter.New(httprouter.NewRegexRouteFactory())

	requestResponseMiddleware := httproutermiddleware.NewRequestResponseLog()
	errorResponseMiddleware := middleware.NewErrorResponse()
	errorLogMiddleware := middleware.NewErrorLog()
	authMiddleware := middleware.NewAuth(userRepository)

	router.Use(requestResponseMiddleware, errorResponseMiddleware, errorLogMiddleware)

	router.Get("/health", &httphandler.Health{})

	router.Group(func(router httprouter.Router) {
		router.Use(authMiddleware)

		router.Post(`/dialog/{user_id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/send`, &handler.SendDialog{
			DialogRepository: dialogRepository,
		})

		router.Get(`/dialog/{user_id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/list`, &handler.ListDialog{
			DialogRepository: dialogRepository,
		})
	})

	httpServer := httpserver.New(httpserver.Config{
		Port:                          envConfig.HTTPPort,
		RequestMaxHeaderBytes:         envConfig.RequestHeaderMaxSize,
		ReadHeaderTimeoutMilliseconds: envConfig.RequestReadHeaderTimeoutMilliseconds,
	}, router)

	httpServerErrCh := httpServer.Start()
	defer httpServer.Shutdown()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case osSignal := <-osSignals:
		slog.Info(fmt.Sprintf("got signal from OS: %v. Exit...", osSignal))
	case err := <-httpServerErrCh:
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

func logLevel(lvl string) slog.Level {
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}
