package database

import (
	"gorm.io/gorm"
)

var (
	DB     *gorm.DB // Master database untuk write operations
	ReadDB *gorm.DB // Replica database untuk read operations
)

// GetWriteDB returns the master database for write operations
func GetWriteDB() *gorm.DB {
	return DB
}

// GetReadDB returns the replica database for read operations
func GetReadDB() *gorm.DB {
	return ReadDB
}

// GetDB returns master database (backward compatibility)
func GetDB() *gorm.DB {
	return DB
}
