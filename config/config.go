package config

import (
	"io"
	"time"

	"github.com/joomcode/errorx"
	"github.com/yookoala/realpath"

	"berty.tech/go-libp2p-tor-transport/internal/confStore"
)

// Check that all configurator are correctly done :
var _ = []Configurator{
	AllowTcpDial,
	DoSlowStart,
	EnableEmbeded,
}

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

// SetBinaryPath set the path to the Tor's binary if you don't use the embeded Tor node.
func SetBinaryPath(path string) Configurator {
	return func(c *confStore.Config) error {
		rpath, err := realpath.Realpath(path)
		if err != nil {
			return errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.ExePath = rpath
		return nil
	}
}

// SetTemporaryDirectory sets the temporary directory where Tor is gonna put his
// data dir.
// If you want an easy way to find it you can use:
// https://github.com/Jorropo/go-temp-dir
func SetTemporaryDirectory(path string) Configurator {
	return func(c *confStore.Config) error {
		rpath, err := realpath.Realpath(path)
		if err != nil {
			errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.TempDataDirBase = rpath
		return nil
	}
}
