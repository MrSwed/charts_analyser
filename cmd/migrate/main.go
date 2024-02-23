package main

import (
	"context"
	"errors"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
)

func main() {
	var (
		db      *sqlx.DB
		err     error
		isNewDB = false
	)
	conf := NewConfig().Init()
	if len(conf.DatabaseDSN) == 0 {
		if conf.DatabaseDSN == "" {
			println("DatabaseDSN is required")
			os.Exit(1)
		}

	}
	if db, err = sqlx.Open("pgx", conf.DatabaseDSN); err != nil {
		log.Printf("cannot connect db %s", err)
	}
	defer db.Close()

	log.Println("DB connected")
	versions, errM := Migrate(conf.MigrateDataPath, db.DB)
	switch {
	case errors.Is(errM, migrate.ErrNoChange):
		log.Println("DB migrate: ", errM, versions)
	case errM == nil:
		log.Println("DB migrate: new applied ", versions)
	default:
		log.Printf("DB migrate: %s %v", errM, versions)
	}
	isNewDB = versions[0] == 0

	if !isNewDB {
		log.Println("not new db, no data import")
		return
	}
	ctx := context.Background()
	err = importZones(ctx, conf.ZonesFile, db)
	if err != nil {
		log.Println(err)
	}

	//go importCharts(conf.ChartsPath, db)

}
