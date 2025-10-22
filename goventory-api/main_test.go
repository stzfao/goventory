package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stzfao/goventory/goventory-api/internal/database"
)

func setupRouter() (*gin.Engine, *sql.DB) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		panic(err)
	}

	server := &Server{db: db}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.POST("/hosts", server.createHostHandler)
	r.GET("/hosts", server.listHostsHandler)
	r.GET("/hosts/:hostname", server.getHostHandler)
	r.PUT("/hosts/:hostname", server.updateHostHandler)
	r.DELETE("/hosts/:hostname", server.deleteHostHandler)

	return r, db
}

func TestCreateHostHandler(t *testing.T) {
	r, db := setupRouter()
	defer db.Close()

	host := database.Host{
		Hostname:  "test-host",
		IPAddress: "1.1.1.1",
		HostGroup: "test-group",
	}
	body, _ := json.Marshal(host)

	req, _ := http.NewRequest(http.MethodPost, "/hosts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdHost database.Host
	json.Unmarshal(w.Body.Bytes(), &createdHost)

	assert.NotEmpty(t, createdHost.ID)
	assert.Equal(t, host.Hostname, createdHost.Hostname)
}

func TestGetHostHandler(t *testing.T) {
	r, db := setupRouter()
	defer db.Close()

	host, _ := database.CreateHost(db, &database.Host{Hostname: "test-host", IPAddress: "1.1.1.1", HostGroup: "test-group"})

	req, _ := http.NewRequest(http.MethodGet, "/hosts/"+host.Hostname, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetchedHost database.Host
	json.Unmarshal(w.Body.Bytes(), &fetchedHost)

	assert.Equal(t, host.ID, fetchedHost.ID)
	assert.Equal(t, host.Hostname, fetchedHost.Hostname)
}

func TestListHostsHandler(t *testing.T) {
	r, db := setupRouter()
	defer db.Close()

	database.CreateHost(db, &database.Host{Hostname: "test-host-1", IPAddress: "1.1.1.1", HostGroup: "test-group"})
	database.CreateHost(db, &database.Host{Hostname: "test-host-2", IPAddress: "2.2.2.2", HostGroup: "test-group"})

	req, _ := http.NewRequest(http.MethodGet, "/hosts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var hosts []database.Host
	json.Unmarshal(w.Body.Bytes(), &hosts)

	assert.Len(t, hosts, 2)
}

func TestUpdateHostHandler(t *testing.T) {
	r, db := setupRouter()
	defer db.Close()

	host, _ := database.CreateHost(db, &database.Host{Hostname: "test-host", IPAddress: "1.1.1.1", HostGroup: "test-group"})

	updateData := map[string]string{
		"ip_address": "1.2.3.4",
		"host_group": "new-group",
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, "/hosts/"+host.Hostname, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedHost database.Host
	json.Unmarshal(w.Body.Bytes(), &updatedHost)

	assert.Equal(t, "1.2.3.4", updatedHost.IPAddress)
	assert.Equal(t, "new-group", updatedHost.HostGroup)
}

func TestDeleteHostHandler(t *testing.T) {
	r, db := setupRouter()
	defer db.Close()

	host, _ := database.CreateHost(db, &database.Host{Hostname: "test-host", IPAddress: "1.1.1.1", HostGroup: "test-group"})

	req, _ := http.NewRequest(http.MethodDelete, "/hosts/"+host.Hostname, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	_, err := database.GetHostByHostname(db, host.Hostname)
	assert.Equal(t, sql.ErrNoRows, err)
}
