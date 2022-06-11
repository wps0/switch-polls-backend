package db

import (
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"switch-polls-backend/config"
)

func initMigrations() *migrate.Migrate {
	workdir, err := os.Getwd()
	dbInstance := OpenDbInstance()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
		return nil
	}
	driver, err := migratemysql.WithInstance(dbInstance, &migratemysql.Config{
		MigrationsTable: TablePrefix + "schema_migrations",
	})
	if err != nil {
		log.Fatalf("Failed to initialise migrate sql driver: %v", err)
		return nil
	}

	migr, err := migrate.NewWithDatabaseInstance("file://"+workdir+string(os.PathSeparator)+config.MigrationsPathRelative, config.Cfg.DatabaseConfig.DBName, driver)
	if err != nil {
		log.Fatalf("Failed to get migrate instance: %v", err)
		return nil
	}
	return migr
}

func ApplyMigrations() {
	log.Println("Applying database migrations...")
	migr := initMigrations()
	defer migr.Close()

	ver, dirty, err := migr.Version()
	if err == migrate.ErrNilVersion {
		log.Println("No migration has ever been applied.")
	} else if err != nil {
		log.Fatalf("Failed to get database version: %v", err)
		return
	} else {
		log.Printf("Current db version: %d (dirtiness: %v)", ver, dirty)
	}

	log.Println("Applying migrations...")
	err = migr.Up()
	if err == migrate.ErrNoChange {
		log.Println("No changes to apply.")
		return
	} else if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
		return
	}

	ver, dirty, err = migr.Version()
	if err != nil {
		log.Fatalf("Failed to get database version: %v", err)
		return
	}
	log.Printf("New db version: %d (dirtiness: %v)", ver, dirty)
	log.Println("Migrations applied successfully.")
}
