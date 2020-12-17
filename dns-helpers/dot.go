package dns

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"time"

	"github.com/joomcode/errorx"
	madns "github.com/multiformats/go-multiaddr-dns"
	"github.com/ncruces/go-dns"
)

func CreatDoTDNSResolverFromDialContext(dialFunc dns.DialFunc, hostname string, addresses ...string) (*net.Resolver, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, errorx.Decorate(err, "can't fetch system cert pool")
	}
	resolver, err := dns.NewDoTResolver(
		hostname,
		dns.DoTAddresses(addresses...),
		dns.DoTCache(
			dns.MaxCacheEntries(256),
			dns.MaxCacheTTL(time.Hour*24),
			dns.MinCacheTTL(time.Minute),
		),
		dns.DoTConfig(&tls.Config{
			RootCAs: certPool,
		}),
		dns.DoTDialFunc(dialFunc),
	)
	if err != nil {
		return nil, errorx.Decorate(err, "can't create DoT resolver")
	}
	return resolver, nil
}

func CreateDoTMaDNSResolverFromDialContext(dialFunc dns.DialFunc, hostname string, addresses ...string) (*madns.Resolver, error) {
	netResolver, err := CreatDoTDNSResolverFromDialContext(dialFunc, hostname, addresses...)
	if err != nil {
		return nil, err
	}
	return &madns.Resolver{
		Backend: netResolver,
	}, nil
}
