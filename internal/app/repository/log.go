package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
)

type LogRepo struct {
	db *sqlx.DB
}

func NewLogRepository(db *sqlx.DB) *LogRepo {
	return &LogRepo{db: db}
}

func (r *LogRepo) ControlLogAdd(ctx context.Context, logs ...domain.ControlLog) (err error) {
	if len(logs) == 0 {
		return
	}
	var tx *sqlx.Tx
	if tx, err = r.db.Beginx(); err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			err = errors.Join(err, rErr)
		}
	}()

	var stmt *sqlx.Stmt
	if stmt, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+constant.DBControlLog+
		" (vessel_id, timestamp, control, comment) VALUES($1, $2, $3, $4)"); err != nil {
		return
	}

	for _, log := range logs {
		if log.Timestamp.IsZero() {
			log.Timestamp = time.Now()
		}
		if _, err = stmt.ExecContext(ctx,
			log.Vessel.ID, log.Timestamp, log.Control, log.Comment); err != nil {
			return
		}
	}
	err = tx.Commit()
	return
}
