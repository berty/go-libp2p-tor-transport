# go-tor-transport
Go tor transport is a [go-libp2p](https://github.com/libp2p/go-libp2p) transport targeting mainly *nix platform.

## WIP
This transport is in very early stages (PoC) many of features enumerated here are just targets we think we can acheive.

## Support
This transport target *nix (Linux, OSX, IOS, Android, ...), it also can be built for windows but many of the attack reduction mesures will not be present (on windows potentialy un encrypted traffic will transit through the loopback, also some tor control port are gonna be exposed.).

## Tor Provider Compatibility
You can use 3 different provider :
- Embeded, this transport naturaly include a tor build, nothing to do.
- External node, the transport will connect to an already running tor node (usefull for server, can increase annonymity by running a relay at the same time blocking SCA attacks).
- File Path, you can give a path to a tor node executable, the transport will start the node and setup it.

## Running Mode
- Compatibility, this will create tunnels only 1 or 0 long, this disable privacy protection but allows to connect or be contacted by annonym nodes at a much lower cost (you can also use that to go around your nat).
- Annonymity, this will works like you expect 3 hop tunnels, in combination with a correct configuration of your libp2p node you will be able to also dial non annonym nodes through an exit relay.

## API
Go-tor-transport implements the standart [`go-libp2p-core/transport#Transport`](https://pkg.go.dev/github.com/libp2p/go-libp2p-core@v0.6.1/transport?tab=doc#Transport) api but the process is not as straigth forward as for simpler transport.

You need to first create a transport builder with `go-tor-transport#NewBuilder` and pass his output to [`go-libp2p#Transport`](https://pkg.go.dev/github.com/libp2p/go-libp2p/?tab=doc#Transport).
`NewBuilder` also accept `go-tor-transport/config#Configurator` as argument to alter the config, you will find them in `go-tor-transport/config`.

## Possible pre 1.0 features :
- DNS Provider, might need some libp2p changes, but currently it would be possible to leak your real ip through the DNS (sending a new domain through the tor connection, libp2p would try to resolve it but not through tor).
- Multiple transports on one tor node (reduce overhead).
- Rebootless config updates (currently to update the config you need to unhook the transport from libp2p (wich is not possible yet) and rebind a new instance with updated config).
