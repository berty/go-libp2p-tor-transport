# go-libp2p-tor-transport
Go tor transport is a [go-libp2p](https://github.com/libp2p/go-libp2p) transport targeting mainly \*nix platform.

You can follow the process of the **1.0** [here](https://github.com/berty/go-libp2p-tor-transport/projects/1).

## Usage :
```go
package main
import (
  "context"

  tor "berty.tech/go-libp2p-tor-transport"
  dns "berty.tech/go-libp2p-tor-transport/dns-helpers"
  config "berty.tech/go-libp2p-tor-transport/config"
  libp2p "github.com/libp2p/go-libp2p"
)

func main() {
  builder, err := tor.NewBuilder( // Create a builder
    config.EnableEmbeded(),       // Use the embeded tor instance.
  )
  c(err)
  host, err := libp2p.New(        // Create a libp2p node
    context.Background(),
    libp2p.Transport(builder.GetTransportConstructor()), // Use the builder to create a transport instance (you can't reuse the same builder after that).
  )
  c(err)
}

func c(err error) { // Used to check error in this example, replace by whatever you want.
  if err != nil {
    panic(err)
  }
}
```

### With config :
```go
package main
import (
  "context"
  "time"

  tor "berty.tech/go-libp2p-tor-transport"
  config "berty.tech/go-libp2p-tor-transport/config"
  libp2p "github.com/libp2p/go-libp2p"
  madns "github.com/multiformats/go-multiaddr-dns"
)

func main() {
  builder, err := tor.NewBuilder(        // NewBuilder can accept some `config.Configurator`
    config.AllowTcpDial(),               // Some Configurator are already ready to use.
    config.SetSetupTimeout(time.Minute), // Some require a parameter, in this case it's a function that will return a Configurator.
    config.SetBinaryPath("/usr/bin/tor"),
  )
  c(err)
  // Sets the default madns resolver, if you don't do that dns requests will be done clearly over internet.
  r, err := dns.CreateDoTMaDNSResolverFromDialContext(
    builder.GetDialer().DialContext,                                      // Dialer
    "cloudflare-dns.com",                                                 // Hostname
    "1.1.1.1", "1.0.0.1", "2606:4700:4700::1111", "2606:4700:4700::1001", // Addresses
  )
  c(err)
  madns.DefaultResolver = r
  // Everything else is as previously shown.
  hostWithConfig, err := libp2p.New(
    context.Background(),
    libp2p.Transport(builder.GetTransportConstructor()),
  )
  c(err)
}

func c(err error) {
  if err != nil {
    panic(err)
  }
}
```

## Support
This transport target \*nix (Linux, OSX, IOS, Android, ...), it also can be built for windows but many of the attack reduction mesures will not be present (on windows potentialy un encrypted traffic will transit through the loopback, also some tor control port are gonna be exposed.).

## Tor Provider Compatibility
You can use 3 different provider :
- Embeded, this transport naturaly include a tor build, nothing to do.
- External node, the transport will connect to an already running tor node (usefull for server, can increase annonymity by running a relay at the same time blocking SCA attacks).
- File Path, you can give a path to a tor node executable, the transport will start the node and setup it.

## Running Mode
- Compatibility, this will create tunnels only 1 or 0 long, this disable privacy protection but allows to connect or be contacted by annonym nodes at a much lower cost (you can also use that to go around your nat) **default**.
- Annonymity, this will works like you expect 3 hop tunnels, in combination with a correct configuration of your libp2p node you will be able to also dial non annonym nodes through an exit relay.
