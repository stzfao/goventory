package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stzfao/goventory/goventory-api/internal/database"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Setup a temporary database for testing
	var err error
	testDB, err = database.NewDB(":memory:")
	if err != nil {
		panic("Failed to create in-memory database: " + err.Error())
	}

	// Run tests
	exitCode := m.Run()

	// Teardown
	testDB.Close()

	os.Exit(exitCode)
}

func setupRouter() *gin.Engine {
	server := &Server{db: testDB}
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong (•̀ᴗ•́ )و"})
	})

	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/hosts", server.createHostHandler)
		apiV1.GET("/hosts", server.listHostsHandler)
		apiV1.GET("/hosts/:hostname", server.getHostHandler)
		apiV1.DELETE("/hosts/:hostname", server.deleteHostHandler)
		apiV1.PUT("/hosts/:hostname", server.updateHostHandler)
	}

	r.GET("/inventory", server.getInventoryHandler)

	return r
}

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"message\":\"pong (•̀ᴗ•́ )و\"}", w.Body.String())
}

func TestHostHandlers(t *testing.T) {
	// Clean up the database before each test
	_, err := testDB.Exec("DELETE FROM hosts")
	assert.NoError(t, err)

	router := setupRouter()

	// Create a host
	host := database.Host{
		Hostname:  "test-host",
		IPAddress: "192.168.1.1",
		HostGroup: "test-group",
	}
	jsonValue, _ := json.Marshal(host)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/hosts", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get the created host
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/hosts/test-host", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var fetchedHost database.Host
	json.Unmarshal(w.Body.Bytes(), &fetchedHost)
	assert.Equal(t, host.Hostname, fetchedHost.Hostname)

	// List hosts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/hosts", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	hosts := []database.Host{}
	json.Unmarshal(w.Body.Bytes(), &hosts)
	assert.Len(t, hosts, 1)
	assert.Equal(t, host.Hostname, hosts[0].Hostname)

	// Update the host
	updatedHost := database.Host{
		IPAddress: "192.168.1.2",
		HostGroup: "new-group",
	}
	jsonValue, _ = json.Marshal(updatedHost)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/hosts/test-host", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Get the updated host
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/hosts/test-host", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var fetchedUpdatedHost database.Host
	json.Unmarshal(w.Body.Bytes(), &fetchedUpdatedHost)
	assert.Equal(t, updatedHost.IPAddress, fetchedUpdatedHost.IPAddress)
	assert.Equal(t, updatedHost.HostGroup, fetchedUpdatedHost.HostGroup)

	// Delete the host
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/hosts/test-host", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify the host is deleted
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/hosts/test-host", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
