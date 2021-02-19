module berty.tech/go-libp2p-tor-transport

go 1.15

require (
	berty.tech/go-libtor v1.0.371
	github.com/cretz/bine v0.1.0
	github.com/joomcode/errorx v1.0.3
	github.com/libp2p/go-libp2p-core v0.8.0
	github.com/libp2p/go-libp2p-transport-upgrader v0.4.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multiaddr-dns v0.2.1-0.20201130213411-dba25a2c0b7a
	github.com/multiformats/go-multiaddr-fmt v0.1.0
	github.com/ncruces/go-dns v1.0.0
	github.com/yookoala/realpath v1.0.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/tools v0.0.0-20200904185747-39188db58858 // indirect
)

replace github.com/ncruces/go-dns => github.com/berty/go-dns v1.0.1 // temporary, see https://github.com/ncruces/go-dns/pull/8/
