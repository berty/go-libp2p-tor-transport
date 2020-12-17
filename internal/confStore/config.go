package confStore

import (
	"context"
	"time"

	"github.com/cretz/bine/tor"
)

// `Config` stores the config, don't use it, you must use Configurator.
type Config struct {
	AllowTcpDial   bool
	SetupTimeout   time.Duration
	RunningContext context.Context

	TorStart *tor.StartConf
}
