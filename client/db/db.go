package db

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

var (
	once          sync.Once
	statusCounter Database
)

type (
	Database interface {
		Save(context.Context, *Data) error
		Get(context.Context, string) (*Data, error)
	}

	databaseService struct {
		db *sql.DB
	}
)

func (d *databaseService) Save(ctx context.Context, data *Data) error {
	panic("implement me")
}

func (d *databaseService) Get(ctx context.Context, s string) (*Data, error) {
	panic("implement me")
}

func GetDB() Database {
	once.Do(func() {
		statusCounter = newDb()
	})
	return statusCounter
}

func newDb() Database {
	db, err := sql.Open("sqlite3", "./data.db")
	checkErr(err)

	return &databaseService{db}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
