package database

import (
	"fmt"

	"github.com/bananocoin/boompow-next/services/server/src/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     string
	Password string
	User     string
	DBName   string
	SSLMode  string
}

func NewConnection(config *Config, mock bool) (*gorm.DB, error) {
	if mock && config.DBName != "testing" {
		panic("Mock flag can only be set for testing database")
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return db, err
	}
	if mock {
		// For mock drop and create
		DropAndCreateTables(db)
	}
	return db, nil
}

func DropAndCreateTables(db *gorm.DB) {
	db.Migrator().DropTable(&models.User{}, &models.WorkRequest{})
	db.Exec(fmt.Sprintf("DROP TYPE %s", models.PG_USER_TYPE_NAME))
	createTypes(db)
	db.Migrator().CreateTable(&models.User{}, &models.WorkRequest{})
}

func Migrate(db *gorm.DB) error {
	createTypes(db)
	return db.AutoMigrate(&models.User{}, &models.WorkRequest{})
}

// Create types in postgres
func createTypes(db *gorm.DB) error {
	result := db.Exec(fmt.Sprintf("SELECT 1 FROM pg_type WHERE typname = '%s';", models.PG_USER_TYPE_NAME))

	switch {
	case result.RowsAffected == 0:
		if err := db.Exec(fmt.Sprintf("CREATE TYPE %s AS ENUM ('%s', '%s');", models.PG_USER_TYPE_NAME, models.PROVIDER, models.REQUESTER)).Error; err != nil {
			fmt.Printf("Error creating %s ENUM", models.PG_USER_TYPE_NAME)
			return err
		}

		return nil
	case result.Error != nil:
		return result.Error

	default:
		return nil
	}
}
