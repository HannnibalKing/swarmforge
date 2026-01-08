package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize router
	router := gin.Default()

	// Middleware
	router.Use(corsMiddleware())
	router.Use(authMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Job routes
		jobs := v1.Group("/jobs")
		{
			jobs.POST("", createJob)
			jobs.GET("/:id", getJob)
			jobs.GET("/:id/parts", getJobParts)
			jobs.PATCH("/:id/status", updateJobStatus)
		}

		// Printer routes
		printers := v1.Group("/printers")
		{
			printers.POST("", registerPrinter)
			printers.GET("/:id", getPrinter)
			printers.GET("/available", getAvailablePrinters)
			printers.POST("/:id/certify", submitCertification)
			printers.PATCH("/:id/status", updatePrinterStatus)
		}

		// QA routes
		qa := v1.Group("/qa")
		{
			qa.POST("/submit", submitQA)
			qa.GET("/pending", getPendingQA)
			qa.POST("/review/:id", reviewQA)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.POST("/register", registerUser)
			users.POST("/login", loginUser)
			users.GET("/me", getCurrentUser)
		}
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public endpoints
		publicPaths := []string{"/health", "/api/v1/users/login", "/api/v1/users/register"}
		for _, path := range publicPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// TODO: Implement JWT validation
		c.Next()
	}
}

// Handler placeholders
func createJob(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create job endpoint"})
}

func getJob(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get job endpoint"})
}

func getJobParts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get job parts endpoint"})
}

func updateJobStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update job status endpoint"})
}

func registerPrinter(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Register printer endpoint"})
}

func getPrinter(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get printer endpoint"})
}

func getAvailablePrinters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get available printers endpoint"})
}

func submitCertification(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Submit certification endpoint"})
}

func updatePrinterStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update printer status endpoint"})
}

func submitQA(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Submit QA endpoint"})
}

func getPendingQA(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get pending QA endpoint"})
}

func reviewQA(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Review QA endpoint"})
}

func registerUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Register user endpoint"})
}

func loginUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Login endpoint"})
}

func getCurrentUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get current user endpoint"})
}
