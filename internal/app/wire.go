//go:build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"sushkov/internal/adapter/postgres"
	"sushkov/internal/handlers"
	"sushkov/internal/interfaces"
	"sushkov/internal/usecase"
)

var userSet = wire.NewSet(
	postgres.NewUserRepo,
	wire.Bind(new(interfaces.UserRepository), new(*postgres.UserRepo)),
	usecase.NewUserUsecase,
	wire.Bind(new(interfaces.UserUsecase), new(*usecase.UserUsecase)),
	handlers.NewUserHandler,
)

func initUserHandler(db *pgxpool.Pool) *handlers.UserHandler {
	wire.Build(userSet)
	return nil
}
