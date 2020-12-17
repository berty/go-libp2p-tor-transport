package config

import (
	"context"
	"io"
	"time"

	"github.com/joomcode/errorx"
	"github.com/yookoala/realpath"

	"berty.tech/go-libp2p-tor-transport/internal/confStore"
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
func AllowTcpDial() Configurator {
	return func(c *confStore.Config) error {
		c.AllowTcpDial = true
		return nil
	}
}

// DoSlowStart set the tor node to bootstrap only when a Dial or a Listen is issued.
// By Default DoSlowStart is off.
func DoSlowStart() Configurator {
	return func(c *confStore.Config) error {
		c.TorStart.EnableNetwork = false
		return nil
	}
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

// SetBinaryPath set the path to the Tor's binary if you don't use the embeded Tor node.
func SetBinaryPath(path string) Configurator {
	rpath, err := realpath.Realpath(path)
	return func(c *confStore.Config) error {
		if err != nil {
			return errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.ExePath = rpath
		return nil
	}
}

// SetTemporaryDirectory sets the temporary directory where Tor is gonna put his
// data dir.
func SetTemporaryDirectory(path string) Configurator {
	rpath, err := realpath.Realpath(path)
	return func(c *confStore.Config) error {
		if err != nil {
			return errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.TempDataDirBase = rpath
		return nil
	}
}

// SetTorrc sets the torrc file for tor to use instead of an blank one.
func SetTorrcPath(path string) Configurator {
	rpath, err := realpath.Realpath(path)
	return func(c *confStore.Config) error {
		if err != nil {
			return errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.TorrcFile = rpath
		return nil
	}
}

// SetRunningContext sets the context used for the running of the node.
func SetRunningContext(ctx context.Context) Configurator {
	return func(c *confStore.Config) error {
		c.RunningContext = ctx
		return nil
	}
}
