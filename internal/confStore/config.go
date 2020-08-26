package confStore

import (
	"time"

	"github.com/cretz/bine/tor"
)

// `Config` stores the config, don't use it, you must use Configurator.
type Config struct {
	AllowTcpDial bool
	SetupTimeout time.Duration

	TorStart *tor.StartConf
}
