package postgres

import (
	"github.com/doug-martin/goqu/v9"
)

type PostgresDictionary struct {
	Goqu *goqu.Database
}
