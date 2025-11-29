terraform {
  required_providers {
    goventory = {
      source  = "registry.terraform.io/stzfao/goventory"
      version = "0.1.0"
    }
  }
}

provider "goventory" {
  server_url = "http://localhost:8080"
}

resource "goventory_host" "test" {
  hostname   = "cli-test-host"
  ip_address = null
  host_group = "test"  # Required field
}