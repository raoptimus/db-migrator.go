package service

import (
	"github.com/pkg/errors"
)

var ErrMigrationVersionReserved = errors.New("migration version reserved")
