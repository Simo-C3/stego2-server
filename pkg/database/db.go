package database

import (
	"database/sql"
	"fmt"

	"github.com/Simo-C3/stego2-server/pkg/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type DB struct {
	*bun.DB
}

func New(cfg *config.DBConfig) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqldb, mysqldialect.New())

	return &DB{db}, nil
}
