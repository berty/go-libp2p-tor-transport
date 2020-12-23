package tor

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
	"time"

	"berty.tech/go-libp2p-tor-transport/config"
	"berty.tech/go-libp2p-tor-transport/internal/confStore"

	"github.com/cretz/bine/tor"

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

	// Used to upgrade unsecure TCP connections to secure multiplexed and
	// authenticate Tor connections.
	upgrader *tptu.Upgrader

	// if allowTcpDial is true the transport will accept to dial tcp address.
	allowTcpDial bool
	// setupTimeout is the timeout for announcing a tunnel
	setupTimeout time.Duration

	// the listenStore is used for dialing connection to exchange listen addr.
	laddrs listenStore
}

// listenStore is a store for listen addrs.
type listenStore struct {
	sync.RWMutex

	cur *listenHolder
}

// listenHolder is a linked list of listen addrs, used to get one to send.
type listenHolder struct {
	addr ma.Multiaddr

	next *listenHolder
}

func NewBuilder(cs ...config.Configurator) (func(*tptu.Upgrader) tpt.Transport, error) {
	var conf confStore.Config
	{
		// Applying configuration
		c := &confStore.Config{
			SetupTimeout: 5 * time.Minute,
			TorStart: &tor.StartConf{
				EnableNetwork: true, // Do Fast Start
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
	return matcher.Matches(maddr) || (t.allowTcpDial && mafmt.TCP.Matches(maddr) && manet.IsPublicAddr(maddr))
}

func (t *transport) Close() {
	t.bridge.Close()
}

// Listen listens on the given multiaddr.
func (t *transport) Listen(laddr ma.Multiaddr) (tpt.Listener, error) {
	var lconf tor.ListenConf
	var base string
	if laddr.Protocols()[0].Code == ma.P_ONION3 {
		lconf = tor.ListenConf{
			// This extract the port from the bytes and add it to the listen
			RemotePorts: []int{int(binary.BigEndian.Uint16(laddr.Bytes()[37:39]))},
			Version3:    true,
		}
		base = "/onion3/"
	} else {
		lconf = tor.ListenConf{
			// This extract the port from the bytes and add it to the listen
			RemotePorts: []int{int(binary.BigEndian.Uint16(laddr.Bytes()[12:14]))},
		}
		base = "/onion/"
	}
	ctx, cancel := context.WithTimeout(context.Background(), t.setupTimeout)
	// Create an onion service to listen on any port but show as 80
	s, err := t.bridge.Listen(ctx, &lconf)
	if err != nil {
		cancel()
		return nil, errorx.Decorate(err, "Can't start tor listener")
	}

	// Adding the listen addr to the store.
	// Making the new holder
	trueLAddr, err := ma.NewMultiaddr(base + s.ID + ":" + strconv.Itoa(s.RemotePorts[0]))
	checkError(err)
	newHolder := &listenHolder{addr: trueLAddr}

	t.laddrs.Lock()
	// Iterating the linked list and adding to the last.
	cur := t.laddrs.cur
	if cur == nil {
		t.laddrs.cur = newHolder
		goto FinishAddingLAddr
	}
	for cur.next != nil {
		cur = cur.next
	}
	cur.next = newHolder
FinishAddingLAddr:
	t.laddrs.Unlock()

	return &listener{
		service:    s,
		ctx:        ctx,
		cancel:     cancel,
		upgrader:   t.upgrader,
		t:          t,
		lAddrStore: &t.laddrs,
		lAddrCur:   newHolder,
	}, nil
}

func (t *transport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.CapableConn, error) {
	// Shouldn't be needed but might avoid ip leaking in case of libp2p bug.
	if !t.CanDial(raddr) {
		return nil, errorx.IllegalArgument.New(fmt.Sprintf("Can't dial \"%s\".", raddr))
	}
	p0 := raddr.Protocols()[0].Code
	switch p0 {
	case ma.P_ONION, ma.P_ONION3: // Onion Dial
		addr := string(maddrToNetAddrP0(raddr, p0))

		// Dialing
		c, err := t.dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, errorx.Decorate(err, "Can't dial")
		}
		// Upgrading
		conn, err := t.upgrader.UpgradeOutbound(ctx, t, &dialConn{
			netConnWithoutAddr: c,
			raddr:              raddr,
			laddr:              &t.laddrs,
		}, p)
		if err != nil {
			return nil, errorx.Decorate(err, "Can't upgrade laddr exchange connection")
		}

		// Entering the laddr exchange (due to tor limitation, the dialer have to send
		// his local addr manualy).
		var laddr ma.Multiaddr
		t.laddrs.RLock()
		cur := t.laddrs.cur
		t.laddrs.RUnlock()
		if cur == nil {
			laddr = NopMaddr2
		} else {
			laddr = cur.addr
		}
		stream, err := conn.OpenStream(ctx)
		if err != nil {
			conn.Close()
			return nil, errorx.Decorate(err, "Can't open laddr exchange stream")
		}
		// buffer the message first, tor isn't fast.
		var buf []byte
		if laddr.Protocols()[0].Code == ma.P_ONION3 {
			buf = []byte{encodeOnion3}
		} else {
			buf = []byte{encodeOnion}
		}
		buf = append(buf, laddr.Bytes()...)
		n := len(buf)
		var done int
		for n > 0 {
			nminus, err := stream.Write(buf[done:])
			if err != nil {
				// In case of error here just end exchange and continue onward.
				goto EndLAddrExchange
			}
			n -= nminus
			done += nminus
		}
	EndLAddrExchange:
		stream.Close()
		return conn, nil
	case ma.P_IP4, ma.P_IP6: // IP Dial
		netAddr, err := manet.ToNetAddr(raddr)
		checkError(err)

		return t.dialTroughProxy(ctx, raddr, netAddr.String(), p)

	case ma.P_DNS4, ma.P_DNS6: // DNS Dial
		domain, err := raddr.ValueForProtocol(p0)
		checkError(err)
		port, err := raddr.ValueForProtocol(ma.P_TCP)
		checkError(err)

		return t.dialTroughProxy(ctx, raddr, domain+".onion:"+port, p)
	default:
		panic(fmt.Sprintf("Was not able to create net Addr from multiaddr, this shouldn't fail, check your multiaddr package or report to maintainers ! (%s)", raddr))
	}
}

func (t *transport) dialTroughProxy(ctx context.Context, raddr ma.Multiaddr, addr string, p peer.ID) (tpt.CapableConn, error) {
	// Dialing
	c, err := t.dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't dial")
	}
	// Upgrading
	conn, err := t.upgrader.UpgradeOutbound(ctx, t, &dialConnTcp{
		netConnWithoutAddr: c,
		raddr:              raddr,
		laddr:              &t.laddrs,
	}, p)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't upgrade connection")
	}
	return conn, nil
}

const (
	encodeOnion  uint8 = 0
	encodeOnion3 uint8 = 1
)
