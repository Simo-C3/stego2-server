package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type DB struct {
	*bun.DB
}

func New(cfg *config.DBConfig) (*DB, error) {
	d, err := cloudsqlconn.NewDialer(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
	}

	mysql.RegisterDialContext("cloudsqlconn",
		func(ctx context.Context, addr string) (net.Conn, error) {
			return d.Dial(ctx, cfg.InstanceConnectionName)
		})
	dsn := fmt.Sprintf("%s:%s@cloudsqlconn(%s:%s)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqldb, mysqldialect.New())

	return &DB{db}, nil
}
