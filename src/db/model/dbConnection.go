package model

import (
	"database/sql"
)

type DBConnection struct {
	DB *sql.DB
}
