package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestAccHostResource(t *testing.T) {
	// Mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "test-id", "hostname": "test-host", "ip_address": "1.1.1.1", "host_group": "test-group"}`))
		case http.MethodGet:
			if r.URL.Path == "/hosts/test-host-updated" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "test-id", "hostname": "test-host-updated", "ip_address": "2.2.2.2", "host_group": "test-group-updated"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "test-id", "hostname": "test-host", "ip_address": "1.1.1.1", "host_group": "test-group"}`))
			}
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "test-id", "hostname": "test-host-updated", "ip_address": "2.2.2.2", "host_group": "test-group-updated"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
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
					hostname   = "test-host-updated"
					ip_address = "2.2.2.2"
					host_group = "test-group-updated"
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("goventory_host.test", "hostname", "test-host-updated"),
					resource.TestCheckResourceAttr("goventory_host.test", "ip_address", "2.2.2.2"),
					resource.TestCheckResourceAttr("goventory_host.test", "host_group", "test-group-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
