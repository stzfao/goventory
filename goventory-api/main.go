package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stzfao/goventory/goventory-api/internal/database"
)

// Server holds the dependencies for our API server.
type Server struct {
	db *sql.DB
}

func main() {
	log.Println("✶⋆.˚ Starting the GoVentory API Server ⊹ ࣪ ˖⭑.ᐟ")

	db, err := database.NewDB("./goventory.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a new Server instance with the database connection.
	server := &Server{db: db}

	r := gin.Default()

	// https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Wire up the handlers
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong (•̀ᴗ•́ )و"})
	})

    // Create new server
	r.POST("/hosts", server.createHostHandler)

	r.Run(":8080") // listen and serve on 8080
}


// Create a new host
func (server *Server) createHostHandler(context *gin.Context) {
	var newHost database.Host

	// Bind the received JSON to newHost to validate it fits the struct
	if err := context.ShouldBindJSON(&newHost); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Access db with server struct 
	createdHost, err := database.CreateHost(server.db, &newHost)
	if err != nil {
		log.Printf("Error creating host: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create host"})
		return
	}

	context.JSON(http.StatusCreated, createdHost)
}