package main

import (
	"charts_analyser/internal/common/semaphore"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
)

func importChart(ctx context.Context, fileName string, db *sqlx.DB) (count uint64, err error) {
	var file *os.File
	file, err = os.Open(fileName)
	if err != nil {
		return
	}
	defer func() {
		if er := file.Close(); er != nil {
			err = errors.Join(err, er)
		}
	}()

	r := csv.NewReader(file)
	// skip headers
	if _, err = r.Read(); err != nil {
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

	var stmtTracks, stmtVessels *sqlx.Stmt
	if stmtTracks, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+DBTracks+
		" (vessel_id,vessel_name,time,location) VALUES($1, $2, $3, ST_GeometryFromText($4))"); err != nil {
		return
	}
	if stmtVessels, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+DBVessels+
		" (id, name) VALUES($1, $2) ON CONFLICT (id) DO nothing"); err != nil {
		return
	}

	for {
		var csvRow []string // timestamp,longitude,latitude,vessel_id,vessel_name
		csvRow, err = r.Read()
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		if err != nil {
			return
		}
		timestamp, er := strconv.ParseInt(csvRow[0], 10, 64)
		if err != nil {
			err = errors.Join(err, er)
		}
		datetime := time.Unix(timestamp, 0)
		coordsStr := "POINT(" + csvRow[1] + " " + csvRow[2] + ")"
		vesselName := csvRow[4]
		vesselID, er := strconv.ParseUint(csvRow[3], 10, 64)
		if er != nil {
			err = errors.Join(err, er)
			continue
		}
		if _, er = stmtTracks.ExecContext(ctx, vesselID, vesselName, datetime, coordsStr); er != nil {
			err = errors.Join(err, er)
			return
		}
		if _, er = stmtVessels.ExecContext(ctx, vesselID, vesselName); er != nil {
			err = errors.Join(err, er)
			return
		}
		count++
	}

	if _, er := tx.ExecContext(ctx, "update vessels set created_at = COALESCE((select min(time) from tracks t where t.vessel_id = vessels.id), vessels.created_at)"); er != nil {
		err = errors.Join(err, er)
	}
	err = tx.Commit()

	return
}

func importCharts(ctx context.Context, path string, db *sqlx.DB) (filesCount, recordsCount uint64) {
	var (
		entries       []os.DirEntry
		wg            sync.WaitGroup
		err           error
		filesCountA   atomic.Uint64
		recordsCountA atomic.Uint64
	)
	entries, err = os.ReadDir(path)
	if err != nil {
		log.Print(err.Error())
		return
	}
	sem := semaphore.New(runtime.NumCPU())
	wg.Add(len(entries))
	for _, e := range entries {
		go func(fileName string) {
			sem.Acquire()
			defer wg.Done()
			defer sem.Release()

			rCount, er := importChart(ctx, path+"/"+fileName, db)
			filesCountA.Add(1)
			recordsCountA.Add(rCount)
			if er != nil {
				log.Printf("%s: imported %d tracks with errors: %s", fileName, rCount, er.Error())
			} else {
				log.Printf("%s: imported %d tracks", fileName, rCount)

			}
		}(e.Name())
	}

	wg.Wait()

	recordsCount = recordsCountA.Load()
	filesCount = filesCountA.Load()
	return
}
