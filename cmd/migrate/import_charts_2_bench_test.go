package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

func dataFileNamesBench(ctx context.Context, path string) (chan string, uint64) {
	filesCh := make(chan string)
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		log.Println("ReadDir error ", err.Error())
	}
	go func() {
		defer close(filesCh)
		for _, f := range dirEntries {
			select {
			case <-ctx.Done():
				return
			case filesCh <- path + "/" + f.Name():
			}
		}
	}()
	return filesCh, uint64(len(dirEntries))
}

func readFileBench(ctx context.Context, fileName string, pageSize int, csvRowsPageCh chan [][]string) {
	csvRowsPage := &[][]string{}
	file, err := os.Open(fileName)
	if err != nil {
		log.Println("file open error: ", err.Error())
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println("file close error: ", err.Error())
			return
		}
	}()

	r := csv.NewReader(file)
	// skip headers
	if _, err = r.Read(); err != nil {
		log.Println("skip headers error: ", err.Error())
		return
	}
	rowsCount := 0
	for {

		select {
		case <-ctx.Done():
			return
		default:
		}
		csvRow, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Printf("file read error %s %s ", fileName, err.Error())
		}

		if len(*csvRowsPage) == pageSize {
			csvRowsPageCh <- *csvRowsPage
			*csvRowsPage = [][]string{} // clear page
		}
		*csvRowsPage = append(*csvRowsPage, csvRow) // add row to current page
		rowsCount++
	}
	// send rest of rows page
	if len(*csvRowsPage) > 0 {
		csvRowsPageCh <- *csvRowsPage
		*csvRowsPage = [][]string{}
	}
	log.Printf("file %s reader processed %d rows", fileName, rowsCount)
}

// readRows
// read each file and send page of csvRow records
func readRowsBench(ctx context.Context, filesCh chan string, pageSize int) chan [][]string {
	csvRowsCh := make(chan [][]string)
	go func() {
		defer close(csvRowsCh)

		for fileName := range filesCh {
			select {
			case <-ctx.Done():
				return
			default:
				readFileBench(ctx, fileName, pageSize, csvRowsCh)
			}
		}
	}()
	return csvRowsCh
}

// readFanOut
// return csv row collection channels
func readFanOutBench(ctx context.Context, filesCh chan string, numWorkers, pageSize int) []chan [][]string {
	csvRowsChs := make([]chan [][]string, numWorkers)
	for i := 0; i < numWorkers; i++ {
		csvRowsChs[i] = readRowsBench(ctx, filesCh, pageSize)
	}

	return csvRowsChs
}

func insertDataBench(ctx context.Context, db *sqlx.DB, csvRows [][]string) (rowInserted uint64) {
	var (
		tx  *sqlx.Tx
		err error
	)
	if tx, err = db.Beginx(); err != nil {
		log.Println("Transaction begin error: ", err.Error())
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			log.Println("Transaction rollback  error: ", rErr.Error())
			rowInserted = 0
		}
	}()

	var stmtTracks, stmtVessels *sqlx.Stmt
	if stmtTracks, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+DBTracks+
		" (vessel_id,time,location) VALUES($1, $2, ST_GeometryFromText($3))"); err != nil {
		return
	}
	if stmtVessels, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+DBVessels+
		" (id, name) VALUES($1, $2) ON CONFLICT (id) DO nothing"); err != nil {
		return
	}

	for _, csvRow := range csvRows {
		if len(csvRow) == 0 {
			continue
		}
		timestamp, err := strconv.ParseInt(csvRow[0], 10, 64)
		if err != nil {
			log.Println("parse timestamp error: ", err.Error())
		}
		datetime := time.Unix(timestamp, 0)
		coordsStr := "POINT(" + csvRow[1] + " " + csvRow[2] + ")"
		vesselName := csvRow[4]
		vesselID, err := strconv.ParseUint(csvRow[3], 10, 64)
		if err != nil {
			log.Println("parse vesselId error: ", err.Error())
			return
		}
		if _, err = stmtTracks.ExecContext(ctx, vesselID, datetime, coordsStr); err != nil {
			log.Println("Insert track error: ", err.Error())
			return
		}
		if _, err = stmtVessels.ExecContext(ctx, vesselID, vesselName); err != nil {
			log.Println("Insert vessel error: ", err.Error())
			return
		}
		rowInserted++
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Transaction commit error: ", err.Error())
	}
	return
}

func insertDBFanInBench(ctx context.Context, db *sqlx.DB, csvRowsChs ...chan [][]string) chan uint64 {
	finalCh := make(chan uint64)
	var wg sync.WaitGroup

	for _, ch := range csvRowsChs {
		wg.Add(1)

		go func(csvRowsCh chan [][]string) {
			defer wg.Done()

			for data := range csvRowsCh {
				select {
				case <-ctx.Done():
					return
				default:
					finalCh <- insertDataBench(ctx, db, data)
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func importChartsPipelineWithWorkers(c context.Context, path string, db *sqlx.DB) (filesCount, recordsCount uint64) {
	ctx, cancel := context.WithCancel(c)
	defer cancel()

	var filesCh chan string

	// stage 1: list of files to file-workers
	filesCh, filesCount = dataFileNamesBench(ctx, path)

	// stage 2: file-workers read each file and send to db-workers
	dataCh := readFanOutBench(ctx, filesCh, runtime.NumCPU(), ImportTransactionSize)

	// stage 3: db-workers insert to db and send num of inserted rows to result channel
	resultCh := insertDBFanInBench(ctx, db, dataCh...)

	// stage 4: get count of success inserted rows
	for res := range resultCh {
		recordsCount += res
	}

	return
}

/*
Import 5024124 tracks from 100 files at 10m2.380681417s
BenchmarkImportChartsPipelineWithWorkers-4   	       1	602380813107 ns/op
*/
func BenchmarkImportChartsPipelineWithWorkers(b *testing.B) {
	var (
		db         *sqlx.DB
		err        error
		benchStart = time.Now()
	)
	conf := NewConfig().WithEnv().CleanParameters()
	if len(conf.DatabaseDSN) == 0 {
		if conf.DatabaseDSN == "" {
			println("DatabaseDSN is required")
			os.Exit(1)
		}
	}
	db, err = sqlx.Open("pgx", conf.DatabaseDSN)
	require.NoError(b, err)

	defer db.Close()

	_, err = db.Exec(`
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS schema_migrations;
DROP TABLE IF EXISTS vessels;
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS control_log;
DROP TABLE IF EXISTS control_dashboard;
DROP TABLE IF EXISTS zones;
`)
	require.NoError(b, err)

	versions, errM := Migrate(conf.MigrateDataPath, db.DB)
	require.NoError(b, errM)
	require.Equal(b, uint(0), versions[0])

	ctx := context.Background()

	log.Println("Start import tracks..")

	var filesCount, recordsCount uint64
	filesCount, recordsCount = importChartsPipelineWithWorkers(ctx, conf.ChartsPath, db)
	log.Printf("Import %d tracks from %d files at %s", recordsCount, filesCount, time.Since(benchStart))
}
