package tor

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/Jorropo/go-tor-transport/config"
	"github.com/Jorropo/go-tor-transport/internal/confStore"

	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"

	"github.com/libp2p/go-libp2p-core/peer"
	tpt "github.com/libp2p/go-libp2p-core/transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"

	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/joomcode/errorx"
)

type transport struct {
	// bridge is the main tor object, he focuses interactions.
	bridge *tor.Tor
	dialer *tor.Dialer

	// Used to upgrade unsecure TCP connections to secure multiplexed.
	// TODO: Stop using default upgrader for tor 2 tor connections.
	upgrader *tptu.Upgrader

	// if allowTcpDial is true the transport will accept to dial tcp address.
	allowTcpDial bool
	// setupTimeout is the timeout for announcing a tunnel
	setupTimeout time.Duration
}

func NewBuilder(cs ...config.Configurator) (func(*tptu.Upgrader) tpt.Transport, error) {
	var conf confStore.Config
	{
		// Applying configuration
		c := &confStore.Config{
			SetupTimeout: 5 * time.Minute,
			TorStart: &tor.StartConf{
				ProcessCreator:         libtor.Creator, // Use Embed
				UseEmbeddedControlConn: true,           // Use Embed conn
				EnableNetwork:          true,           // Do Fast Start
			},
		}
		if err := config.Merge(cs...)(c); err != nil {
			return nil, errorx.Decorate(err, "Can't apply configuration to the tor node")
		}
		conf = *c
	}
	t, err := tor.Start(context.Background(), conf.TorStart)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't start tor node")
	}

	// Up until this point, we don't need the starting configuration anymore.
	conf.TorStart = nil

	dialer, err := t.Dialer(context.Background(), nil)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't create a dialer.")
	}
	return func(u *tptu.Upgrader) tpt.Transport {
		return &transport{
			allowTcpDial: conf.AllowTcpDial,
			setupTimeout: conf.SetupTimeout,
			bridge:       t,
			dialer:       dialer,
			upgrader:     u,
		}
	}, nil
}

func (_ *transport) Proxy() bool {
	return false
}

var supportedProtos = []int{ma.P_ONION3, ma.P_ONION}
var supportedProtosWithTcp = append(supportedProtos, ma.P_TCP)

func (t *transport) Protocols() []int {
	if t.allowTcpDial {
		return supportedProtosWithTcp
	}
	return supportedProtos
}

var matcher = mafmt.Or(
	mafmt.Base(ma.P_ONION3),
	mafmt.Base(ma.P_ONION),
)

func (t *transport) CanDial(maddr ma.Multiaddr) bool {
	return matcher.Matches(maddr) || (t.allowTcpDial && mafmt.TCP.Matches(maddr))
}

func (t *transport) Close() {
	t.bridge.Close()
}

// Listen listens on the given multiaddr.
func (t *transport) Listen(laddr ma.Multiaddr) (tpt.Listener, error) {
	var lconf tor.ListenConf
	if laddr.Protocols()[0].Code == ma.P_ONION3 {
		lconf = tor.ListenConf{
			// This extract the port from the bytes and add it to the listen
			RemotePorts: []int{int(binary.BigEndian.Uint16(laddr.Bytes()[36:38]))},
			Version3:    true,
		}
	} else {
		lconf = tor.ListenConf{
			// This extract the port from the bytes and add it to the listen
			RemotePorts: []int{int(binary.BigEndian.Uint16(laddr.Bytes()[11:13]))},
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), t.setupTimeout)
	// Create an onion service to listen on any port but show as 80
	s, err := t.bridge.Listen(ctx, &lconf)
	if err != nil {
		cancel()
		return nil, errorx.Decorate(err, "Can't start tor listener")
	}
	return t.upgrader.UpgradeListener(t, &listener{
		service: s,
		cancel:  cancel,
	}), nil
}

func (t *transport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.CapableConn, error) {
	// Shouldn't be needed but might avoid ip leaking in case of libp2p bug.
	if !t.CanDial(raddr) {
		return nil, errorx.IllegalArgument.New(fmt.Sprintf("Can't dial \"%s\".", raddr))
	}
	var addr string
	p0 := raddr.Protocols()[0].Code
	switch p0 {
	case ma.P_ONION, ma.P_ONION3:
		v, err := raddr.ValueForProtocol(p0)
		checkError(err)
		vs := strings.SplitN(v, ":", 2)
		addr = vs[0] + ".onion:" + vs[1]
	case ma.P_IP4, ma.P_IP6:
		n, err := manet.ToNetAddr(raddr)
		checkError(err)
		addr = n.String()
	default:
		panic(fmt.Sprintf("Was not able to create net Addr from multiaddr, this shouldn't fail, check your multiaddr package or report to maintainers ! (%s)", raddr))
	}
	c, err := t.dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't dial")
	}
	return t.upgrader.UpgradeOutbound(ctx, t, &dialConn{
		Conn:  c,
		raddr: raddr,
	}, p)
}
