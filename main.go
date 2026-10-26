package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"os"
	"stock/authorize"
	"stock/db"
	"stock/myRedis"
	"stock/stock"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("No .env file found, proceeding with environment variables")
		} else {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}
	newApp := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port if not specified
	}

	connectionString := os.Getenv("DB_CONNECTION_STRING")
	if connectionString == "" {
		connectionString = "user:password@tcp(localhost:3306)/dbname" // default connection string if not specified
	}
	redisHost := os.Getenv("REDIS_ADDR")
	if redisHost == "" {
		redisHost = "localhost:3003"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = "password0"
	}

	secretKey := os.Getenv("SESSION_SECRET")
	if secretKey == "" {
		slog.Error("can not found secret key app is down")
		os.Exit(1)
	}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	mysql.RegisterTLSConfig("custom-tls", tlsConfig)
	mysqlConnection, err := sqlx.Open("mysql", connectionString)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		os.Exit(1)
	}
	defer mysqlConnection.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mysqlConnection.PingContext(ctx); err != nil {
		log.Printf("Error pinging database: %v", err)
		os.Exit(1)
	}

	pemString, err := os.ReadFile("public.pem")
	if err != nil {
		log.Fatalf("Error reading public key: %v", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pemString)
	if err != nil {
		log.Fatalf("Error parsing public key: %v", err)
	}

	myPool := &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", redisHost, redigo.DialPassword(redisPassword), redigo.DialUseTLS(true), redigo.DialConnectTimeout(3*time.Second))
		},
	}
	mySes := myRedis.NewPrefixRedisStoreWithPool(myPool, "stockService", 3600, []byte(secretKey))

	stockPemString, err := os.ReadFile("stock_public.pem")
	if err != nil {
		log.Fatalf("Error reading stock public key: %v", err)
	}
	stockPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(stockPemString)
	if err != nil {
		log.Fatalf("Error parsing stock public key: %v", err)
	}

	apiKey := os.Getenv("FACILITY_API_KEY")

	apiUrl := os.Getenv("FACILITY_API_URL")

	userDB := db.NewUserDB(mysqlConnection)
	authHandler := authorize.NewAuthHandler(publicKey, userDB, mySes, apiKey, apiUrl, stockPublicKey)

	newApp.Use(authorize.CheckAccessToken(authHandler))
	newApp.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	stockRepository := stock.NewStockRepository(userDB, mySes)
	stockHandler := stock.NewStockHandler(stockRepository)
	newApp.GET("/stock/:facility_id/:room_type_id/:month/:day", stockHandler.GetPerMonth)
	newApp.POST("/stock/stock/:facility_id/:room_type_id/:year/:month/:day/:value", stockHandler.UpdateStock)
	newApp.POST(("/stock/reserved/:facility_id/:room_type_id/:year/:month/:day/:value"), stockHandler.UpdateReserved)
	newApp.POST("/stock/stopSell/:facility_id/:room_type_id/:year/:month/:day/:value", stockHandler.UpdateStopSell)
	newApp.Run(":" + port)
	slog.Info("runnning application in " + port)
}
