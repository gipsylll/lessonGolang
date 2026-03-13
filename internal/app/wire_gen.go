//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"sushkov/internal/adapter/postgres"
	"sushkov/internal/handlers"
	"sushkov/internal/interfaces"
	"sushkov/internal/usecase"
)

func initUserHandler(db *pgxpool.Pool) *handlers.UserHandler {
	userRepo := postgres.NewUserRepo(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handlers.NewUserHandler(userUsecase)
	return userHandler
}

var userSet = wire.NewSet(postgres.NewUserRepo, wire.Bind(new(interfaces.UserRepository), new(*postgres.UserRepo)), usecase.NewUserUsecase, wire.Bind(new(interfaces.UserUsecase), new(*usecase.UserUsecase)), handlers.NewUserHandler)

var _ = userSet
