package databases

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBpools *pgxpool.Pool

func DBPGX(ctx context.Context, dbUrl string) error {
	dbpool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return err
	}
	if err := dbpool.Ping(ctx); err != nil {
		return err
	}

	DBpools = dbpool

	return nil
}
