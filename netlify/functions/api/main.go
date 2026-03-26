package main

import (
	"log"
	"os"

	"absensholat-api/routes"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// 1. Initialize Zap logger
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	sugar := logger.Sugar()

	// 2. Initialize Database Connection
	conn := os.Getenv("DATABASE_URL")
	if conn == "" {
		sugar.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{})
	if err != nil {
		sugar.Fatal("Failed to connect to database:", err)
	}

	// 3. Initialize Shared Gin Engine
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	router := routes.SetupEngine(db, sugar, isProduction)

	// 4. Hook into AWS Lambda Gin Adapter
	ginLambda = ginadapter.New(router)
}

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Proxy AWS Lambda Proxy Request through Gin Engine
	return ginLambda.Proxy(req)
}

func main() {
	// Start the Lambda handler
	lambda.Start(Handler)
}
