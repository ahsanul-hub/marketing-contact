package database

import (
	"app/config"
	"app/dto/model"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"

	// "github.com/golang-migrate/migrate/v4/database/postgres"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var MongoClient *mongo.Client

func ConnectDB() *gorm.DB {
	var err error

	// Construct the Data Source Name (DSN) for PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Config("DB_HOST", ""),
		config.Config("DB_USER", ""),
		config.Config("DB_PASSWORD", ""),
		config.Config("DB_NAME", ""),
		config.Config("DB_PORT", "5432"))

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	// dsnUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	// 	config.Config("DB_USER", ""),
	// 	config.Config("DB_PASSWORD", ""),
	// 	config.Config("DB_HOST", "localhost"),
	// 	config.Config("DB_PORT", "5432"),
	// 	config.Config("DB_NAME", ""),
	// )

	// fmt.Println("dsn: ", dsnUrl)

	// err = runMigrations(dsnUrl)
	// if err != nil {
	// 	log.Fatalf("Failed to run migrations: %v", err)
	// }

	// Get the underlying sql.DB object to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database object: %v", err)
	}

	sqlDB.SetMaxOpenConns(60)

	sqlDB.SetMaxIdleConns(20)

	err = DB.AutoMigrate(&model.User{}, &model.Client{}, &model.ClientApp{}, &model.Transactions{}, &model.PaymentMethodClient{}, &model.PaymentMethod{}, &model.SettlementClient{}, &model.BlockedMDN{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database Migrated")
	return DB
}

// func SetupMongoDB() {
// 	uri := config.Config("MONGODB_URI", "")
// 	var err error
// 	MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
// 	if err != nil {
// 		panic("failed to connect to MongoDB")
// 	}

//		log.Println("Connected to MongoDB")
//	}

func runMigrations(dsn string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrationsPath := filepath.Join(cwd, "migrations")
	m, err := migrate.New(
		migrationsPath,
		dsn,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrated successfully")
	return nil
}

func GetCollection(databaseName, collectionName string) *mongo.Collection {
	return MongoClient.Database(databaseName).Collection(collectionName)
}
