package database

import (
	"fmt"
	"guycanella-url-shortner/internal/config"
	"guycanella-url-shortner/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Sao_Paulo",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	var gormLogger logger.Interface

	if cfg.App.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to obtain SQL DB instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db

	log.Println("✅ Connected to the PostgreSQL database.")
	return nil
}

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("Database not initialized.")
	}

	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("⚠️  Warning: it was not possible to create extension uuid-ossp: %v", err)
	}

	if err := DB.AutoMigrate(&models.URL{}); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	log.Println("✅ Migrations executed successfully.")
	return nil
}

func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
