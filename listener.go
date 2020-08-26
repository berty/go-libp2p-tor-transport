package tor

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cretz/bine/tor"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/joomcode/errorx"
)

type listener struct {
	service *tor.OnionService
	cancel  func()
}

// checkError is used to check if an error ocured during the build of multiaddr, if this fail that mean you have a problem with your packages.
func checkError(err error) {
	if err != nil {
		panic(fmt.Sprintf("Was not able to create onion multiaddr, this shouldn't fail, check your multiaddr package or report to maintainers ! (%s)", err))
	}
}

func (l *listener) Multiaddr() ma.Multiaddr {
	if l.service.Version3 {
		m, err := ma.NewMultiaddr("/onion3/" + l.service.ID + ":" + strconv.Itoa(l.service.RemotePorts[0]))
		checkError(err)
		return m

	}
	m, err := ma.NewMultiaddr("/onion/" + l.service.ID + ":" + strconv.Itoa(l.service.RemotePorts[0]))
	checkError(err)
	return m
}

func (l *listener) Addr() net.Addr {
	return l.service.Addr()
}

func (l *listener) Close() error {
	l.cancel()
	return l.service.Close()
}

func (l *listener) Accept() (manet.Conn, error) {
	c, err := l.service.Accept()
	if err != nil {
		return nil, errorx.Decorate(err, "Can't accept connection")
	}
	return &listConn{
		Conn: c,
		l:    l,
	}, nil
}
