package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath, migrationsTable, databaseDSN string
	var downFlag bool
	var verbose bool

	flag.StringVar(
		&databaseDSN,
		"d",
		"",
		"database-dsn",
	)
	flag.StringVar(
		&migrationsPath,
		"p",
		"",
		"path to migrations",
	)
	flag.StringVar(
		&migrationsTable,
		"t",
		"migrations",
		"name of migration table, where migrator writes own data",
	)
	flag.BoolVar(
		&downFlag,
		"down",
		false,
		"run down migrations instead of up",
	)
	flag.BoolVar(
		&verbose,
		"v",
		false,
		"verbose output",
	)
	flag.Parse()

	if databaseDSN == "" {
		fmt.Println("Error: database DSN is required")
		os.Exit(1)
	}

	if migrationsPath == "" {
		fmt.Println("Error: migrations path is required")
		os.Exit(1)
	}

	fileInfo, err := os.Stat(migrationsPath)
	if err != nil {
		fmt.Printf("Error accessing migrations path: %v\n", err)
		os.Exit(1)
	}

	if !fileInfo.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", migrationsPath)
		os.Exit(1)
	}

	if verbose {
		files, err := os.ReadDir(migrationsPath)
		if err != nil {
			fmt.Printf("Error reading migrations directory: %v\n", err)
		} else {
			fmt.Println("Files in migrations directory:")
			for _, file := range files {
				fmt.Printf("  %s\n", file.Name())
			}
		}
	}

	migrationSource := "file://" + migrationsPath
	if verbose {
		fmt.Printf("Using migration source: %s\n", migrationSource)
		fmt.Printf("Using database DSN: %s\n", databaseDSN)
	}

	m, err := migrate.New(
		migrationSource,
		databaseDSN+"&x-migrations-table="+migrationsTable,
	)
	if err != nil {
		fmt.Printf("Error creating migrator: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			fmt.Printf("Error closing source: %v\n", srcErr)
		}
		if dbErr != nil {
			fmt.Printf("Error closing DB: %v\n", dbErr)
		}
	}()

	if downFlag {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to revert")
				return
			}
			fmt.Printf("Error applying down migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migrations reverted successfully")
	} else {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}
			fmt.Printf("Error applying up migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migrations applied successfully")
	}
}
