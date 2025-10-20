package db

import (
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var mu sync.Mutex

func InitDB(path string, migrate bool) (*gorm.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	if DB != nil {
		return DB, nil
	}
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Prevent unnecessary overhead from bridges
	// Only HTTPServer should have migrate=true
	if migrate {
		err = database.AutoMigrate(&Project{})
		if err != nil {
			return nil, err
		}
	}

	DB = database
	return DB, nil
}

func GetDB() *gorm.DB {
	if DB == nil {
		panic("database not initialized yet")
	}
	return DB
}
