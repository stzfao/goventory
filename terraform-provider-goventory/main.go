package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/stzfao/goventory/terraform-provider-goventory/internal/provider"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	err := providerserver.Serve(context.Background(), provider.New("dev"), providerserver.ServeOpts{
		Address: "hashicorp.com/edu/goventory",
	})

	if err != nil {
		log.Fatal(err.Error())
	}
}