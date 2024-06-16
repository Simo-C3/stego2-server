package infra

import (
	"context"
	"database/sql"
	"log"

	"github.com/pkg/errors"

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

func convertToDomainProblem(problem *ProblemModel) *model.Problem {
	return &model.Problem{
		ID:              problem.ID,
		CollectSentence: problem.CollectSentence,
		Level:           problem.Level,
	}
}

// GetProblems implements repository.ProblemRepository.
func (p *problemRepository) GetProblems(ctx context.Context, level, limit int) ([]*model.Problem, error) {
	problems := make([]ProblemModel, 0, limit)
	query := p.db.NewSelect().
		Model(&problems).
		// Where("? <= level", level-1).
		// Where("level <= ?", level+1).
		Where("level BETWEEN ? AND ?", level-1, level+1).
		OrderExpr("RAND()").
		Limit(limit)

	log.Println(query.String())
	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := make([]*model.Problem, 0, len(problems))
	for _, problem := range problems {
		res = append(res, convertToDomainProblem(&problem))
	}

	return res, nil
}
