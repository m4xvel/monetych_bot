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

type GameWithTypeRow struct {
	GameID   int
	GameName string
	TypeID   *int
	TypeName *string
}

type GameRepository interface {
	Get(ctx context.Context) ([]GameWithTypeRow, error)
}
