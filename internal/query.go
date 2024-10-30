package internal

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbInstance struct {
	db *pgxpool.Pool
}

func (service dbInstance) QueryDB(sqlQuery string) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := service.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (service dbInstance) QueryCount(sqlQuery string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var count int
	err := service.db.QueryRow(ctx, sqlQuery).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func QueryTournamentsList(fields *string, pool *pgxpool.Pool) error {
	sqlQuery := `SELECT DISTINCT(match_type) FROM event`
	dbV := dbInstance{db: pool} // I know this is not correct, will work on this
	rows, err := dbV.QueryDB(sqlQuery)
	if err != nil {
		return err
	}

	// if it is multiple, create a slice here and append to it and finally set it to fields
	for rows.Next() {
		err = rows.Scan(fields)
		if err != nil {
			return err
		}
	}
	defer rows.Close()
	return nil
}
