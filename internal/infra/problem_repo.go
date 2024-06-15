package infra

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/uptrace/bun"
)

type ProblemModel struct {
	bun.BaseModel `bun:"table:problems"`

	ID              int    `bun:",pk"`
	CollectSentence string `bun:"collect_sentence"`
	Level           int    `bun:"level"`
}

type problemRepository struct {
	db *database.DB
}

func NewProblemRepository(db *database.DB) repository.ProblemRepository {
	return &problemRepository{
		db: db,
	}
}

// func convertToProblemModel(problem *model.Problem) *ProblemModel {
// 	return &ProblemModel{
// 		ID:              problem.ID,
// 		CollectSentence: problem.CollectSentence,
// 		Level:           problem.Level,
// 	}
// }

func convertToDomainProblem(problem *ProblemModel) *model.Problem {
	return &model.Problem{
		ID:              problem.ID,
		CollectSentence: problem.CollectSentence,
		Level:           problem.Level,
	}
}

// GetProblems implements repository.ProblemRepository.
func (p *problemRepository) GetProblems(ctx context.Context, level int) (*model.Problem, error) {
	var problem ProblemModel
	query := p.db.NewSelect().
		Model(&problem).
		Where("level BETWEEN ? AND ?", level-1, level+1).
		OrderExpr("RAND()").
		Limit(1)
	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return convertToDomainProblem(&problem), nil
}
