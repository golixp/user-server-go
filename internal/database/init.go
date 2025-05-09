// Package database provides database client initialization.
package database

import (
	"strings"
	"sync"

	"github.com/go-dev-frame/sponge/pkg/sgorm"

	"user-server-go/internal/config"
)

var (
	gdb     *sgorm.DB
	gdbOnce sync.Once

	ErrRecordNotFound = sgorm.ErrRecordNotFound
)

// InitDB connect database
func InitDB() {
	dbDriver := config.Get().Database.Driver
	switch strings.ToLower(dbDriver) {
	case sgorm.DBDriverMysql, sgorm.DBDriverTidb:
		gdb = InitMysql()
	case sgorm.DBDriverPostgresql:
		gdb = InitPostgresql()
	case sgorm.DBDriverSqlite:
		gdb = InitSqlite()
	default:
		panic("InitDB error, please modify the correct 'database' configuration at yaml file. " +
			"Refer to https://github.com/golixp/user-server-go/blob/main/configs/user_server_go.yml#L41")
	}
}

// GetDB get db
func GetDB() *sgorm.DB {
	if gdb == nil {
		gdbOnce.Do(func() {
			InitDB()
		})
	}

	return gdb
}

// CloseDB close db
func CloseDB() error {
	return sgorm.CloseDB(gdb)
}
