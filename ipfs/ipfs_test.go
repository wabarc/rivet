package ipfs

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner"
)

func TestLocally(t *testing.T) {
	host, port := "localhost", 5001
	_, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Skipf("ipfs is not running: %v", err)
	}

	opts := []PinningOption{
		Mode(Local),
		Host(host),
		Port(port),
	}

	p := Options(opts...)
	b := []byte(helper.RandString(6, "lower"))
	_, err = (&Locally{p}).Pin(b)
	if err != nil {
		t.Errorf("Unexpected pin data locally: %v", err)
	}
}

func TestRemotely(t *testing.T) {
	opts := []PinningOption{
		Mode(Remote),
		Uses(pinner.Infura),
	}

	p := Options(opts...)
	b := []byte(helper.RandString(6, "lower"))
	_, err := (&Remotely{p}).Pin(b)
	if err != nil {
		t.Errorf("Unexpected pin data remotely: %v", err)
	}
}
