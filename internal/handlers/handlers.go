package handlers

import (
	"database/sql"
)

type Handlers struct {
	db *sql.DB
}

func New(db *sql.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}
