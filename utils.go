package tor

import (
	"fmt"
	"net"
	"strings"

	ma "github.com/multiformats/go-multiaddr"
)

// nopMaddr is an empty maddr used as filler when needed.
var nopMaddr, _ = ma.NewMultiaddr("/onion/aaaaaaaaaaaaaaaa:1")
var nopOnion3, _ = ma.NewMultiaddr("/onion3/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1")

// maddrToNetAddr is unsafe and need to be called with caution.
func maddrToNetAddr(m ma.Multiaddr) addr {
	return maddrToNetAddrP0(m, m.Protocols()[0].Code)
}

func maddrToNetAddrP0(m ma.Multiaddr, p0 int) addr {
	v, err := m.ValueForProtocol(p0)
	checkError(err)
	vs := strings.SplitN(v, ":", 2)
	return addr(vs[0] + ".onion:" + vs[1])
}

// Used to build bine friendly tor net.Addr.
type addr string

var _ net.Addr = addr("")

func (_ addr) Network() string {
	return "tcp"
}

func (a addr) String() string {
	return string(a)
}

// checkError is used to check if an error ocured during the build of multiaddr, if this fail that mean you have a problem with your packages.
func checkError(err error) {
	if err != nil {
		panic(fmt.Sprintf("Was not able to create onion multiaddr, this shouldn't fail, check your multiaddr package or report to maintainers ! (%s)", err))
	}
}
