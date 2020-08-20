// Package main implements the oasis-registry binary which provides tooling to manage a filesystem
// based Oasis Metadata Registry.
package main

import (
	"github.com/hashicorp/go-plugin"

	"github.com/oasisprotocol/metadata-registry-tools/oasis-registry/cmd"
)

func main() {
	// If we use go-plugin, we are supposed to clean clients up.
	defer plugin.CleanupClients()

	cmd.Execute()
}
