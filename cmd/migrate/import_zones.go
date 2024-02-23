package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"strconv"
	"strings"
)

type zones map[string][][]float64

func importZones(ctx context.Context, file string, db *sqlx.DB) (count uint, err error) {
	var data []byte

	if data, err = os.ReadFile(file); err != nil {
		log.Println(err)
		return
	}

	var zonesData zones
	if err = json.Unmarshal(data, &zonesData); err != nil {
		return
	}

	var tx *sqlx.Tx
	if tx, err = db.Beginx(); err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			err = errors.Join(err, rErr)
		}
	}()

	var stmt *sqlx.Stmt
	if stmt, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+DBZones+
		" (name, geometry) VALUES($1, ST_GeometryFromText($2))"); err != nil {
		return
	}

	for name, coords := range zonesData {
		// проверка на замкнутость полигона
		if coords[0][0] != coords[len(coords)-1][0] || coords[0][1] != coords[len(coords)-1][1] {
			coords = append(coords, coords[0])
		}

		var pcoords []string
		for _, c := range coords {
			str := strconv.FormatFloat(c[0], 'g', -1, 64) + " " +
				strconv.FormatFloat(c[1], 'g', -1, 64)
			pcoords = append(pcoords, str)
		}
		coordsStr := "polygon((" + strings.Join(pcoords, ",") + "))"
		if _, err = stmt.ExecContext(ctx, name, coordsStr); err != nil {
			return
		}
		count++
	}
	err = tx.Commit()

	return
}
