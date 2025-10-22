terraform {
  required_providers {
    goventory = {
      source  = "hashicorp.com/edu/goventory"
      version = "dev"
    }
  }
}

provider "goventory" {
  server_url = "http://localhost:8080"
  # you can switch this with your host IP
}

resource "goventory_host" "web1" {
  hostname   = "BMI-XXXXX"
  ip_address = "192.168.1.10"
  host_group = "test-webservers"
}