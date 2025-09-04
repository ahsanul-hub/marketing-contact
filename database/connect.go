package database

import (
	"app/config"
	"app/dto/model"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"

	// "github.com/golang-migrate/migrate/v4/database/postgres"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var MongoClient *mongo.Client

func SetupSQLLogfile() io.Writer {
	logDir := "../logs/sql"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	currentTime := time.Now()
	year, month, _ := currentTime.Date()
	_, week := currentTime.ISOWeek()

	logFilename := filepath.Join(logDir,
		fmt.Sprintf("dcb-new-sql-%d-%02d-week%d.log", year, month, week))

	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Logging initialized")

	return logFile
}

func ConnectDB() *gorm.DB {
	var err error

	if logWriter == nil {
		logWriter = SetupSQLLogfile()
	}

	newLogger := logger.New(
		log.New(logWriter, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             1000 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Construct the Data Source Name (DSN) for Master PostgreSQL
	masterDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Config("DB_HOST", ""),
		config.Config("DB_USER", ""),
		config.Config("DB_PASSWORD", ""),
		config.Config("DB_NAME", ""),
		config.Config("DB_PORT", "5432"))

	// Connect to Master Database
	DB, err = gorm.Open(postgres.Open(masterDSN), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect to master database")
	}

	// Construct the Data Source Name (DSN) for Replica PostgreSQL
	replicaDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Config("DB_REPLICA_HOST", config.Config("DB_HOST", "")), // Fallback ke master jika tidak ada replica
		config.Config("DB_REPLICA_USER", config.Config("DB_USER", "")),
		config.Config("DB_REPLICA_PASSWORD", config.Config("DB_PASSWORD", "")),
		config.Config("DB_REPLICA_NAME", config.Config("DB_NAME", "")),
		config.Config("DB_REPLICA_PORT", config.Config("DB_PORT", "5432")))

	// Connect to Replica Database
	ReadDB, err = gorm.Open(postgres.Open(replicaDSN), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to replica database, using master for reads: %v", err)
		ReadDB = DB // Fallback ke master database jika replica tidak tersedia
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

	// Configure connection pool for Master Database
	masterSqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get master database object: %v", err)
	}
	masterSqlDB.SetMaxOpenConns(20)
	masterSqlDB.SetMaxIdleConns(6)

	// Configure connection pool for Replica Database
	replicaSqlDB, err := ReadDB.DB()
	if err != nil {
		log.Fatalf("Failed to get replica database object: %v", err)
	}
	replicaSqlDB.SetMaxOpenConns(30) // Lebih banyak koneksi untuk read operations
	replicaSqlDB.SetMaxIdleConns(10)

	err = DB.AutoMigrate(&model.User{}, &model.Client{}, &model.ClientApp{}, &model.Transactions{}, &model.PaymentMethodClient{}, &model.PaymentMethod{}, &model.SettlementClient{}, &model.BlockedMDN{}, &model.BlockedUserId{}, &model.ChannelRouteWeight{})
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
