package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getDSN() string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		"aws-0-eu-central-1.pooler.supabase.com",
		"postgres.hzqptekopyduwhazytud",
		"JRUkRf4hYqe0uxLM",
		"postgres",
		"6543",
	)
	log.Printf("Attempting to connect to database with DSN: %s", dsn)
	return dsn
}

func InitDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(getDSN()), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	if err := configureConnectionPool(db); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")

	// if viper.GetString("node.mode") == "development" {
	err = AutoMigrate(db)
	if err != nil {
		log.Printf("Failed to auto migrate: %v", err)
		return nil, err
	}
	// }

	return db, nil
}
