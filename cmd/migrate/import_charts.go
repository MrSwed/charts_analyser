package main

import (
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
	"time"

	"github.com/jmoiron/sqlx"
)

func dataFileNames(ctx context.Context, path string) (chan string, uint64) {
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

func readFile(ctx context.Context, fileName string, pageSize int, csvRowsPageCh chan [][]string) {
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
func readRows(ctx context.Context, filesCh chan string, pageSize int) chan [][]string {
	csvRowsCh := make(chan [][]string)
	go func() {
		defer close(csvRowsCh)

		for fileName := range filesCh {
			select {
			case <-ctx.Done():
				return
			default:
				readFile(ctx, fileName, pageSize, csvRowsCh)
			}
		}
	}()
	return csvRowsCh
}

// readFanOut
// return csv row collection channels
func readFanOut(ctx context.Context, filesCh chan string, numWorkers, pageSize int) []chan [][]string {
	csvRowsChs := make([]chan [][]string, numWorkers)
	for i := 0; i < numWorkers; i++ {
		csvRowsChs[i] = readRows(ctx, filesCh, pageSize)
	}

	return csvRowsChs
}

func insertData(ctx context.Context, db *sqlx.DB, csvRows [][]string) (rowInserted uint64) {
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
		if err != nil {
			return
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
			if err != nil {
				log.Println("Insert vessel error: ", err.Error())
			}
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

func insertDBFanIn(ctx context.Context, db *sqlx.DB, csvRowsChs ...chan [][]string) chan uint64 {
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
					finalCh <- insertData(ctx, db, data)
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

func importCharts(c context.Context, path string, db *sqlx.DB) (filesCount, recordsCount uint64) {
	ctx, cancel := context.WithCancel(c)
	defer cancel()

	var filesCh chan string

	// stage 1: list of files to file-workers
	filesCh, filesCount = dataFileNames(ctx, path)

	// stage 2: file-workers read each file and send to db-workers
	dataCh := readFanOut(ctx, filesCh, runtime.NumCPU(), ImportTransactionSize)

	// stage 3: db-workers insert to db and send num of inserted rows to result channel
	resultCh := insertDBFanIn(ctx, db, dataCh...)

	// stage 4: get count of success inserted rows
	for res := range resultCh {
		recordsCount += res
	}

	return
}

func finishImportMigrate(ctx context.Context, db *sqlx.DB) {
	if _, err := db.ExecContext(ctx, "update "+DBVessels+" set created_at = COALESCE((select min(time) from "+DBTracks+" t where t.vessel_id = vessels.id), vessels.created_at)"); err != nil {
		log.Print(err.Error())
	}

	if _, err := db.ExecContext(ctx, "select setval('"+DBVessels+"_id_seq', (SELECT MAX(id) FROM "+DBVessels+"))"); err != nil {
		log.Print(err.Error())
	}
}
