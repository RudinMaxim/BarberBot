package database

import (
	"fmt"
	"log"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"gorm.io/gorm"
)

type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Version   int    `gorm:"uniqueIndex"`
	Name      string `gorm:"uniqueIndex"`
	AppliedAt time.Time
}

var migrations = []struct {
	Version int
	Name    string
	Up      func(*gorm.DB) error
	Down    func(*gorm.DB) error
}{
	{
		Version: 1,
		Name:    "create_client_table",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&common.Client{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&common.Client{})
		},
	},
	{
		Version: 1,
		Name:    "create_service_table",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&common.Service{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&common.Service{})
		},
	},
	{
		Version: 1,
		Name:    "create_working_hours_table",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&common.WorkingHours{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&common.WorkingHours{})
		},
	},
	{
		Version: 1,
		Name:    "create_appointment_table",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&common.Appointment{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&common.Appointment{})
		},
	},
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running auto migration")
	if err := RunMigrations(db); err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}
	log.Println("Auto migration completed successfully")
	return nil
}

func InitMigrationTable(db *gorm.DB) error {
	return db.AutoMigrate(&Migration{})
}

func RunMigrations(db *gorm.DB) error {
	if err := InitMigrationTable(db); err != nil {
		return fmt.Errorf("failed to initialize migration table: %v", err)
	}

	for _, migration := range migrations {
		var m Migration
		if err := db.Where("version = ?", migration.Version).First(&m).Error; err == gorm.ErrRecordNotFound {
			log.Printf("Applying migration %d: %s", migration.Version, migration.Name)
			if err := migration.Up(db); err != nil {
				return fmt.Errorf("failed to apply migration %d (%s): %v", migration.Version, migration.Name, err)
			}
			db.Create(&Migration{Version: migration.Version, Name: migration.Name, AppliedAt: time.Now()})
			log.Printf("Successfully applied migration %d: %s", migration.Version, migration.Name)
		}
	}

	return nil
}

func RollbackLastMigration(db *gorm.DB) error {
	var lastMigration Migration
	if err := db.Order("version DESC").First(&lastMigration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %v", err)
	}
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].Version == lastMigration.Version {
			log.Printf("Rolling back migration %d: %s", lastMigration.Version, lastMigration.Name)
			if err := migrations[i].Down(db); err != nil {
				return fmt.Errorf("failed to rollback migration %d (%s): %v", lastMigration.Version, lastMigration.Name, err)
			}
			if err := db.Delete(&lastMigration).Error; err != nil {
				return fmt.Errorf("failed to delete migration record: %v", err)
			}
			log.Printf("Successfully rolled back migration %d: %s", lastMigration.Version, lastMigration.Name)
			return nil
		}
	}

	return fmt.Errorf("migration %d not found in migration list", lastMigration.Version)
}
