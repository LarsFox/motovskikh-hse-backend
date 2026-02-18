package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env из корня проекта (на уровень выше migrations/)
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Подключение к БД
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot ping database:", err)
	}

	fmt.Println("✅ Connected to database")

	// Создаём таблицу для отслеживания миграций
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Failed to create schema_migrations table:", err)
	}

	// Получаем текущую версию
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		log.Fatal("Failed to get current version:", err)
	}

	fmt.Printf("Current schema version: %d\n", currentVersion)

	// Читаем файлы миграций
	files, err := filepath.Glob("*.sql")
	if err != nil {
		log.Fatal("Failed to read migration files:", err)
	}
	sort.Strings(files)

	// Применяем новые миграции
	applied := 0
	for _, file := range files {
		var version int
		fmt.Sscanf(filepath.Base(file), "%d_", &version)

		if version <= currentVersion {
			continue
		}

		fmt.Printf("Applying migration %s...\n", filepath.Base(file))

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", file, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Migration %s failed: %v", file, err)
		}

		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			log.Fatalf("Failed to record migration %d: %v", version, err)
		}

		fmt.Printf("✅ Migration %d applied\n", version)
		applied++
	}

	if applied == 0 {
		fmt.Println("No new migrations to apply")
	} else {
		fmt.Printf("\n🎉 Successfully applied %d migration(s)!\n", applied)
	}
}
