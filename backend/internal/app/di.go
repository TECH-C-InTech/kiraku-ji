package app

import (
	"fmt"

	"backend/internal/adapter/repository/memory"
	"backend/internal/port/repository"
	drawusecase "backend/internal/usecase/draw"
)

// Container は API で使用する依存を保持する。
type Container struct {
	DrawFortuneUsecase *drawusecase.FortuneUsecase
}

// NewContainer は依存を初期化して返す。
func NewContainer() (*Container, error) {
	repo, err := provideDrawRepository()
	if err != nil {
		return nil, fmt.Errorf("provide draw repository: %w", err)
	}

	usecase := drawusecase.NewFortuneUsecase(repo)

	return &Container{
		DrawFortuneUsecase: usecase,
	}, nil
}

func provideDrawRepository() (repository.DrawRepository, error) {
	return memory.NewInMemoryDrawRepository(), nil
}
