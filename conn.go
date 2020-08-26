package tor

import (
	"net"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type listConn struct {
	net.Conn

	l *listener
}

func (c *listConn) LocalMultiaddr() ma.Multiaddr {
	return c.l.Multiaddr()
}

func (c *listConn) RemoteMultiaddr() ma.Multiaddr {
	// TODO: implement safe unleakable `conn.RemoteAddr`.
	m, err := manet.FromNetAddr(c.RemoteAddr())
	checkError(err)
	return m
}

type dialConn struct {
	net.Conn

	raddr ma.Multiaddr
}

func (c *dialConn) RemoteMultiaddr() ma.Multiaddr {
	return c.raddr
}

func (c *dialConn) LocalMultiaddr() ma.Multiaddr {
	// TODO: implement safe unleakable `conn.LocalAddr`.
	m, err := manet.FromNetAddr(c.LocalAddr())
	checkError(err)
	return m
}
