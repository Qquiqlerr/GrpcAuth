package migrator

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
)

func main() {
	var storagePath, migrationsPath string
	flag.StringVar(&storagePath, "storage-path", "", "Path to a database storage directory")
	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to a migrations directory")
	flag.Parse()
	if storagePath == "" || migrationsPath == "" {
		panic("storagePath, migrationsPath are required")
	}
	m, err := migrate.New("file://"+migrationsPath, storagePath)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			panic(err)
		}
	}
	fmt.Println("Migrations successfully migrated")
}
