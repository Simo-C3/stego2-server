package repository

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
)

type ProblemRepository interface {
	GetProblems(ctx context.Context, level, limit int) ([]*model.Problem, error)
}
