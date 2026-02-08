package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/handler"
	"github.com/gigaonion/taskalyst/backend/internal/infra/db"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/gigaonion/taskalyst/backend/pkg/customvalidator"
	"github.com/go-playground/validator/v10"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	repo := repository.New(pool)
	txManager := db.NewTxManager(pool)
	userUsecase := usecase.NewUserUsecase(repo, txManager, cfg)
	userHandler := handler.NewUserHandler(userUsecase)
	projectUsecase := usecase.NewProjectUsecase(repo)
	projectHandler := handler.NewProjectHandler(projectUsecase)
	taskUsecase := usecase.NewTaskUsecase(repo)
	taskHandler := handler.NewTaskHandler(taskUsecase)
	timeUsecase := usecase.NewTimeUsecase(repo)
	timeHandler := handler.NewTimeHandler(timeUsecase)
	apiTokenHandler := handler.NewApiTokenHandler(repo)
	calendarUsecase := usecase.NewCalendarUsecase(repo)
	calendarHandler := handler.NewCalendarHandler(calendarUsecase)
	caldavUsecase := usecase.NewCalDavUsecase(repo)
	caldavHandler := handler.NewCalDavHandler(caldavUsecase, calendarUsecase)
	resultUsecase := usecase.NewResultUsecase(repo)
	resultHandler := handler.NewResultHandler(resultUsecase)

	e := echo.New()
	e.Validator = &customvalidator.CustomValidator{Validator: validator.New()} //バリデータを登録
	handler.RegisterRoutes(e, userHandler, projectHandler, taskHandler, timeHandler, apiTokenHandler, cfg, calendarHandler, resultHandler, caldavHandler, repo)

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS()) // 開発用

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// 終了シグナル
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// タイムアウト
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
