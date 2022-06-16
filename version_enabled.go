//go:build embedTor
// +build embedTor

package tor

import (
	"berty.tech/go-libtor"
)

// ProviderVersion return the version of the embeded tor node.
func ProviderVersion() string {
	return libtor.ProviderVersion()
}
