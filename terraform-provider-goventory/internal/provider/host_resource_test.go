package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
	  provider "goventory" {
		server_url = "%s"
	  }
	`
)

var (
	protoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"goventory": providerserver.NewProtocol6WithError(New("test")()),
	}
)

// mockAPIServer is a stateful mock API server for testing.
type mockAPIServer struct {
	mu    sync.Mutex
	hosts map[string]goventoryHost
}

func newMockAPIServer() *mockAPIServer {
	return &mockAPIServer{
		hosts: make(map[string]goventoryHost),
	}
}

func (s *mockAPIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) < 3 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "hosts" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hostname := ""
	if len(parts) > 3 {
		hostname = parts[3]
	}

	switch r.Method {
	case http.MethodPost:
		var host goventoryHost
		if err := json.NewDecoder(r.Body).Decode(&host); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		host.ID = uuid.New().String()
		s.hosts[host.Hostname] = host
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(host)

	case http.MethodGet:
		if hostname == "" {
			var hosts []goventoryHost
			for _, h := range s.hosts {
				hosts = append(hosts, h)
			}
			json.NewEncoder(w).Encode(hosts)
			return
		}
		host, ok := s.hosts[hostname]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(host)

	case http.MethodPut:
		existingHost, ok := s.hosts[hostname]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ip, ok := updates["ip_address"]; ok {
			existingHost.IPAddress = ip.(string)
		}
		if group, ok := updates["host_group"]; ok {
			existingHost.HostGroup = group.(string)
		}
		s.hosts[hostname] = existingHost
		json.NewEncoder(w).Encode(existingHost)

	case http.MethodDelete:
		if _, ok := s.hosts[hostname]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		delete(s.hosts, hostname)
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func TestAccHostResource(t *testing.T) {
	server := httptest.NewServer(newMockAPIServer())
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(providerConfig, server.URL) + `
				  resource "goventory_host" "test" {
					hostname   = "test-host"
					ip_address = "1.1.1.1"
					host_group = "test-group"
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("goventory_host.test", "hostname", "test-host"),
					resource.TestCheckResourceAttr("goventory_host.test", "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("goventory_host.test", "host_group", "test-group"),
					resource.TestCheckResourceAttrSet("goventory_host.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(providerConfig, server.URL) + `
				  resource "goventory_host" "test" {
					hostname   = "test-host"
					ip_address = "2.2.2.2"
					host_group = "test-group-updated"
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("goventory_host.test", "hostname", "test-host"),
					resource.TestCheckResourceAttr("goventory_host.test", "ip_address", "2.2.2.2"),
					resource.TestCheckResourceAttr("goventory_host.test", "host_group", "test-group-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
