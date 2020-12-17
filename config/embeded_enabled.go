// +build embedTor

package config

import (
	"fmt"

	"berty.tech/go-libtor"

	"berty.tech/go-libp2p-tor-transport/internal/confStore"
)

// EnableEmbeded setups the node to use a builtin tor node.
// Note: you will need to build with `-tags=embedTor` for this to works.
// Not available on all systems.
func EnableEmbeded() Configurator {
	return func(c *confStore.Config) error {
		if libtor.Available {
			c.TorStart.ProcessCreator = libtor.Creator
			c.TorStart.UseEmbeddedControlConn = true
			return nil
		}
		return fmt.Errorf("The embeded Tor node isn't available. Check your arch and CGO.")
	}
}
