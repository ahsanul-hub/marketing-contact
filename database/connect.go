package database

import (
	"app/config"
	"app/dto/model"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var MongoClient *mongo.Client

func ConnectDB() {
	var err error

	// Construct the Data Source Name (DSN) for PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Config("DB_HOST", "localhost"),
		config.Config("DB_USER", ""),
		config.Config("DB_PASSWORD", ""),
		config.Config("DB_NAME", ""),
		config.Config("DB_PORT", "5432"))

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect to database")
	}

	fmt.Println("Connection Opened to Database")

	// Migrasi model yang diinginkan
	err = DB.AutoMigrate(&model.User{}, &model.Client{}, &model.Transactions{}, &model.PaymentMethodClient{}, &model.PaymentMethod{}, &model.SettlementClient{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database Migrated")
}

func SetupMongoDB() {
	uri := config.Config("MONGODB_URI", "")
	var err error
	MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic("failed to connect to MongoDB")
	}

	log.Println("Connected to MongoDB")
}

func GetCollection(databaseName, collectionName string) *mongo.Collection {
	return MongoClient.Database(databaseName).Collection(collectionName)
}
