//go:build !embedTor
// +build !embedTor

package config

import (
	"fmt"

	"berty.tech/go-libp2p-tor-transport/internal/confStore"
)

// EnableEmbeded setups the node to use a builtin tor node.
// Note: you will need to build with `-tags=embedTor` for this to works.
// Not available on all systems.
func EnableEmbeded(c *confStore.Config) error {
	return fmt.Errorf("You havn't enabled the embeded tor instance at compilation. You can enable it with `-tags=embedTor` while building.")
}
