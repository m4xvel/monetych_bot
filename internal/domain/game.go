package domain

import "context"

type Game struct {
	ID   int
	Name string
}

type GameType struct {
	ID   int
	Name string
}

type GameRepository interface {
	GetAll(ctx context.Context) ([]Game, error)
	GetTypes(ctx context.Context, gameID int) ([]string, error)
}
