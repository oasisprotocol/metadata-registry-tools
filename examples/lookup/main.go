package main

import (
	"context"
	"fmt"

	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"

	registry "github.com/oasisprotocol/metadata-registry-tools"
)

func main() {
	// Create a new instance of the registry client, pointing to the production instance.
	// Note that in order to refresh the data you currently need to create a new provider.
	gp, err := registry.NewGitProvider(registry.NewGitConfig())
	if err != nil {
		fmt.Printf("Failed to create Git registry provider: %s\n", err)
		return
	}

	ctx := context.Background()

	// Get a list of all entities in the registy.
	entities, err := gp.GetEntities(ctx)
	if err != nil {
		fmt.Printf("Failed to get a list of entities in registy: %s\n", err)
		return
	}

	var entityID signature.PublicKey
	for id, meta := range entities {
		fmt.Printf("[%s]\n", id)
		fmt.Printf("  Name:    %s\n", meta.Name)
		fmt.Printf("  URL:     %s\n", meta.URL)
		fmt.Printf("  Email:   %s\n", meta.Email)
		fmt.Printf("  Keybase: %s\n", meta.Keybase)
		fmt.Printf("  Twitter: %s\n", meta.Twitter)
		fmt.Printf("\n")

		entityID = id
	}

	// Get metadata for a specific entity.
	if len(entities) > 0 {
		entity, err := gp.GetEntity(ctx, entityID)
		if err != nil {
			fmt.Printf("Failed to get entity %s: %s\n", entityID, err)
			return
		}

		fmt.Printf("Random entity:\n")
		fmt.Printf("%+v\n", entity)
	}
}
