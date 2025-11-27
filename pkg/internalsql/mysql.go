package internalsql

import (
	"database/sql"
	"fmt"
	"go-tweets/internal/config"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectMySQL(cfg *config.Config) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, "Asia%2FJakarta")

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database")
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// --- 🔥 Connection Pool Settings ---
	db.SetMaxOpenConns(10)           // max connection concurrency
	db.SetMaxIdleConns(5)            // idle connections allowed
	db.SetConnMaxLifetime(time.Hour) // reuse connection max lifetime

	log.Println("database connected")

	return db, nil
}
