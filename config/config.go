package config

import (
	"io"
	"time"

	"github.com/berty/go-tor-transport/internal/confStore"

	"github.com/joomcode/errorx"
)

type Configurator func(*confStore.Config) error

// ConfMerge Merges different configs, starting at the first ending at the last.
func Merge(cs ...Configurator) Configurator {
	return func(c *confStore.Config) error {
		for _, v := range cs {
			if err := v(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// AllowTcpDial allows the tor transport to dial tcp address.
// By Default TcpDial is off.
func AllowTcpDial(c *confStore.Config) error {
	c.AllowTcpDial = true
	return nil
}

// DoSlowStart set the tor node to bootstrap only when a Dial or a Listen is issued.
// By Default DoSlowStart is off.
func DoSlowStart(c *confStore.Config) error {
	c.TorStart.EnableNetwork = false
	return nil
}

// SetSetupTimeout change the timeout for the bootstrap of the node and the publication of the tunnel.
// By Default SetupTimeout is at 5 minutes.
func SetSetupTimeout(t time.Duration) Configurator {
	return func(c *confStore.Config) error {
		if t == 0 {
			return errorx.IllegalArgument.New("Timeout can't be 0.")
		}
		c.SetupTimeout = t
		return nil
	}
}

// SetNodeDebug set the writer for the tor node debug output.
func SetNodeDebug(debug io.Writer) Configurator {
	return func(c *confStore.Config) error {
		c.TorStart.DebugWriter = debug
		return nil
	}
}
