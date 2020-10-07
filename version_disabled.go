// +build !embedTor

package tor

// ProviderVersion return the version of the embeded tor node.
func ProviderVersion() string {
	return "embeded Disabled"
}
