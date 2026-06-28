package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

type Config struct {
}

func main() {
	if err := loadEnv(); err != nil {
		log.Fatal("Error loading .env file")
	}

	MYSQL_USER := os.Getenv("MYSQL_USER")
	MYSQL_PASSWORD := os.Getenv("MYSQL_PASSWORD")
	MYSQL_HOST := os.Getenv("MYSQL_HOST")
	MYSQL_PORT := os.Getenv("MYSQL_PORT")
	MYSQL_DBNAME := os.Getenv("MYSQL_DBNAME")

	fmt.Println("Starting migration...")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true", MYSQL_USER, MYSQL_PASSWORD, MYSQL_HOST, MYSQL_PORT, MYSQL_DBNAME)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping failed: %v", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("mysql.WithInstance failed: %v", err)
	}

	m, _ := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)

	arg := os.Args[1]
	switch arg {
	case "up":
		err = m.Up()
		if err != nil {
			log.Fatalf("m.Up failed: %v", err)
		}

		fmt.Println("\n--------------------------------")
		fmt.Println("Migration up completed successfully")
		fmt.Println("\n--------------------------------")
	case "down":
		err = m.Steps(-1)
		if err != nil {
			log.Fatalf("m.Down failed: %v", err)
		}
		fmt.Println("\n--------------------------------")
		fmt.Println("Migration down completed successfully")
		fmt.Println("\n--------------------------------")
	case "force":
		version := os.Args[2]

		var v int
		_, err := fmt.Sscanf(version, "%d", &v)
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}

		err = m.Force(v)
		if err != nil {
			log.Fatalf("m.Force failed: %v", err)
		}

		fmt.Println("\n--------------------------------")
		fmt.Println("Migration force completed successfully")
		fmt.Println("\n--------------------------------")
	}

}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}
