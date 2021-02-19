package tor

import (
	"io"
	"net"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
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
		laddr = NopMaddr2
	} else {
		laddr = cur.addr
	}
	return laddr
}

type dialConnTcp struct {
	netConnWithoutAddr

	laddr *listenStore
	raddr ma.Multiaddr
}

func (c *dialConnTcp) LocalAddr() net.Addr {
	return maddrToNetAddr(c.LocalMultiaddr())
}

func (c *dialConnTcp) RemoteAddr() net.Addr {
	addr, err := manet.ToNetAddr(c.RemoteMultiaddr())
	checkError(err)
	return addr
}

func (c *dialConnTcp) RemoteMultiaddr() ma.Multiaddr {
	return c.raddr
}

func (c *dialConnTcp) LocalMultiaddr() ma.Multiaddr {
	var laddr ma.Multiaddr
	c.laddr.RLock()
	cur := c.laddr.cur
	c.laddr.RUnlock()
	if cur == nil {
		laddr = NopMaddr2
	} else {
		laddr = cur.addr
	}
	return laddr
}

// netConnWithoutAddr is a net.Conn like but without LocalAddr and RemoteAddr.
type netConnWithoutAddr interface {
	io.ReadWriteCloser
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}
