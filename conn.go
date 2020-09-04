package tor

import (
	"net"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

type listConn struct {
	netConnWithoutAddr

	l     *listener
	raddr ma.Multiaddr
}

func (c *listConn) LocalAddr() net.Addr {
	return maddrToNetAddr(c.LocalMultiaddr())
}

func (c *listConn) RemoteAddr() net.Addr {
	return maddrToNetAddr(c.RemoteMultiaddr())
}

func (c *listConn) LocalMultiaddr() ma.Multiaddr {
	return c.l.Multiaddr()
}

func (c *listConn) RemoteMultiaddr() ma.Multiaddr {
	return c.raddr
}

type dialConn struct {
	netConnWithoutAddr

	laddr *listenStore
	raddr ma.Multiaddr
}

func (c *dialConn) LocalAddr() net.Addr {
	return maddrToNetAddr(c.LocalMultiaddr())
}

func (c *dialConn) RemoteAddr() net.Addr {
	return maddrToNetAddr(c.RemoteMultiaddr())
}

func (c *dialConn) RemoteMultiaddr() ma.Multiaddr {
	return c.raddr
}

func (c *dialConn) LocalMultiaddr() ma.Multiaddr {
	var laddr ma.Multiaddr
	c.laddr.RLock()
	cur := c.laddr.cur
	c.laddr.RUnlock()
	if cur == nil {
		laddr = nopMaddr
	} else {
		laddr = cur.addr
	}
	return laddr
}

// netConnWithoutAddr is a net.Conn like but without LocalAddr and RemoteAddr.
type netConnWithoutAddr interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}
