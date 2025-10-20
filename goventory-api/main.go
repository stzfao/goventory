package main

import (
	"database/sql"
	"log"
	"net/http"
    // "strconv"

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

    // List existing servers
	r.GET("/hosts", server.listHostsHandler)

    r.GET("/hosts/:id", server.getHostHandler)

    r.DELETE("/hosts/:id", server.deleteHostHandler)

    r.PUT("/hosts/:id", server.updateHostHandler)

	r.Run(":8080") // listen and serve on 8080
}


// Create a new host
func (server *Server) createHostHandler(context *gin.Context) {
	var newHost database.Host

	// Bind the received JSON to newHost to validate it fits the struct
	if err := context.ShouldBindJSON(&newHost); err != nil {
        log.Printf("Data not in expected format: %v", err)
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

// List all hosts
func (server *Server) listHostsHandler(context *gin.Context) {
	hosts, err := database.ListHosts(server.db)
	if err != nil {
		log.Printf("Error listing hosts: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve hosts"})
		return
	}

	context.JSON(http.StatusOK, hosts)
}

// Get a single host by its ID
func (server *Server) getHostHandler(context *gin.Context) {
	id := context.Param("id")
	// id, err := strconv.ParseInt(idStr, 10, 64)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
	// 	return
	// }

	host, err := database.GetHostByID(server.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			context.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			log.Printf("Error getting host %d: %v", id, err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve host"})
		}
		return
	}

	context.JSON(http.StatusOK, host)
}

// Delete a host by its ID
func (server *Server) deleteHostHandler(context *gin.Context) {
	id := context.Param("id")
	// id, err := strconv.ParseInt(idStr, 10, 64)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
	// 	return
	// }

	err := database.DeleteHostByID(server.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			context.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			log.Printf("Error deleting host %d: %v", id, err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete host"})
		}
		return
	}

	context.Status(http.StatusNoContent)
}

// Update a host by its ID
func (server *Server) updateHostHandler(context *gin.Context) {
	id := context.Param("id")
	// id, err := strconv.ParseInt(idStr, 10, 64)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
	// 	return
	// }

	// fetch existing host data
	existingHost, err := database.GetHostByID(server.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			context.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			log.Printf("Error getting host %d: %v", id, err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve host"})
		}
		return
	}

	// bind json to a map for partial updates
	var updates map[string]interface{}
	if err := context.ShouldBindJSON(&updates); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// merge
	if hostname, ok := updates["hostname"].(string); ok {
		existingHost.Hostname = hostname
	}
	if ipAddress, ok := updates["ip_address"].(string); ok {
		existingHost.IPAddress = ipAddress
	}
	if hostGroup, ok := updates["host_group"].(string); ok {
		existingHost.HostGroup = hostGroup
	}

	// Save the updated host data
	err = database.UpdateHost(server.db, existingHost)
	if err != nil {
		// The update function already checks for ErrNoRows, but it's good practice to be robust.
		if err == sql.ErrNoRows {
			context.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			log.Printf("Error updating host %s: %v", id, err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update host"})
		}
		return
	}

	context.JSON(http.StatusOK, existingHost)
}
