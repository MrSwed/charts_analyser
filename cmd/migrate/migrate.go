package main

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

//go:generate migrate create -ext sql -dir ../../migrate -format 20060102030405 charts_analyser

func Migrate(fileSource string, db *sql.DB) (version [2]uint, err error) {
	var driver database.Driver
	if driver, err = postgres.WithInstance(db, &postgres.Config{}); err != nil {
		return
	}
	var m *migrate.Migrate
	if m, err = migrate.NewWithDatabaseInstance(fileSource, "", driver); err != nil {
		return
	}
	version[0], _, _ = m.Version()
	err = m.Up()
	version[1], _, _ = m.Version()
	return
}
