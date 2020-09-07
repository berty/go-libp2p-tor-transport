package tor

import (
	"sync"
	"testing"
  "runtime"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	tpt "github.com/libp2p/go-libp2p-core/transport"
	mplex "github.com/libp2p/go-libp2p-mplex"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"

	ttransport "github.com/libp2p/go-libp2p-testing/suites/transport"

	ma "github.com/multiformats/go-multiaddr"
)

var nopOnion3, _ = ma.NewMultiaddr("/onion3/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1")
var maddrTest = []ma.Multiaddr{
	nopMaddr,
	nopOnion3,
}

func TestTransport(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(len(maddrTest) * len(ttransport.Subtests))
	for _, maddr := range maddrTest {
		for _, suite := range ttransport.Subtests {
			ia := makeInsecureTransport(t)
			ib := makeInsecureTransport(t)

			ba, err := NewBuilder()
			if err != nil {
				t.Fatalf("Can't create builder : %q", err)
			}
			ta := ba(&tptu.Upgrader{
				Secure: ia,
				Muxer:  new(mplex.Transport),
			})
			bb, err := NewBuilder()
			if err != nil {
				t.Fatalf("Can't create builder : %q", err)
			}
			tb := bb(&tptu.Upgrader{
				Secure: ib,
				Muxer:  new(mplex.Transport),
			})
			func(f func(*testing.T, tpt.Transport, tpt.Transport, ma.Multiaddr, peer.ID), ft *testing.T, fta, ftb tpt.Transport, zero ma.Multiaddr, pid peer.ID) {
				defer wg.Done()
				f(ft, fta, ftb, zero, pid)
			}(suite, t, ta, tb, maddr, ia.LocalPeer())
		}
	}
	wg.Wait()
}

func makeInsecureTransport(t *testing.T) *insecure.Transport {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	if err != nil {
		t.Fatal(err)
	}
	id, err := peer.IDFromPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return insecure.NewWithIdentity(id, priv)
}
