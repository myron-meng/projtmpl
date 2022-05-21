package repository

import (
	"database/sql"
)

type Repositories struct {
}

func New(db *sql.DB) *Repositories {
	return &Repositories{}
}
