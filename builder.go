package tor

import (
	"context"
	"net"
	"time"

	"github.com/cretz/bine/tor"

	"github.com/joomcode/errorx"
	"golang.org/x/net/proxy"

	"berty.tech/go-libp2p-tor-transport/config"
	"berty.tech/go-libp2p-tor-transport/internal/confStore"
)

// Builder is the type holding the starter node, it's used to fetch different ressources.
type Builder struct {
	allowTcpDial bool
	setupTimeout time.Duration
	bridge       *tor.Tor
	dialer       ContextDialer
}

// ContextDialer is a dialler that also support contexted dials.
type ContextDialer interface {
	proxy.Dialer
	DialContext(ctx context.Context, network string, addr string) (net.Conn, error)
}

func NewBuilder(cs ...config.Configurator) (*Builder, error) {
	var conf confStore.Config
	{
		// Applying configuration
		conf = confStore.Config{
			SetupTimeout:   5 * time.Minute,
			RunningContext: context.Background(),
			TorStart: &tor.StartConf{
				EnableNetwork: true, // Do Fast Start
			},
		}
		if err := config.Merge(cs...)(&conf); err != nil {
			return nil, errorx.Decorate(err, "Can't apply configuration to the tor node")
		}
	}
	t, err := tor.Start(conf.RunningContext, conf.TorStart)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't start tor node")
	}

	// Up until this point, we don't need the starting configuration anymore.
	conf.TorStart = nil

	dialer, err := t.Dialer(conf.RunningContext, nil)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't create a dialer.")
	}
	return &Builder{
		allowTcpDial: conf.AllowTcpDial,
		setupTimeout: conf.SetupTimeout,
		bridge:       t,
		dialer:       dialer,
	}, nil
}

// GetDialer returns a shared dialer, it is closed once the transport closes.
func (b *Builder) GetDialer() ContextDialer {
	return b.dialer
}
