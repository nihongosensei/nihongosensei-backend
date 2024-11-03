package postgres

import (
	"database/sql"
	"github.com/doug-martin/goqu/v9"
)

type PostgresDictionary struct {
	DB   *sql.DB
	Goqu *goqu.Database
}
