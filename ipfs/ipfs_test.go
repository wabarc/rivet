package ipfs

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner"
)

var (
	ipfsCid = "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
	addJSON = fmt.Sprintf(`{
  "Bytes": 0,
  "Hash": "%s",
  "Name": "name",
  "Size": "1B"
}`, ipfsCid)
)

func TestLocally(t *testing.T) {
	handleResponse := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/add":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(addJSON))
		}
	}

	_, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	opts := []PinningOption{
		Mode(Local),
		Host(host),
		Port(port),
	}

	p := Options(opts...)
	b := []byte(helper.RandString(6, "lower"))
	i, err := (&Locally{p}).Pin(b)
	if err != nil {
		t.Errorf("Unexpected pin data locally: %v", err)
	}
	if i != ipfsCid {
		t.Fatalf("Unexpected cid got %s instead of %s", i, ipfsCid)
	}
}

func TestRemotely(t *testing.T) {
	apikey := os.Getenv("IPFS_PINNER_PINATA_API_KEY")
	secret := os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
	if apikey == "" || secret == "" {
		t.Skip(`Must set env "IPFS_PINNER_PINATA_API_KEY" and "IPFS_PINNER_PINATA_SECRET_API_KEY"`)
	}

	opts := []PinningOption{
		Mode(Remote),
		Uses(pinner.Pinata),
		Apikey(apikey),
		Secret(secret),
	}

	p := Options(opts...)
	b := []byte(helper.RandString(6, "lower"))
	_, err := (&Remotely{p}).Pin(b)
	if err != nil {
		t.Errorf("Unexpected pin data remotely: %v", err)
	}
}

func TestRateLimit(t *testing.T) {
	counter := 0
	handleResponse := func(w http.ResponseWriter, r *http.Request) {
		counter++
		if counter <= maxRetries {
			_, _ = w.Write([]byte(``))
			return
		}
		switch r.URL.Path {
		case "/api/v0/add":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(addJSON))
		}
	}

	_, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	opts := []PinningOption{
		Mode(Local),
		Host(host),
		Port(port),
		Backoff(true),
	}

	p := Options(opts...)
	b := []byte(helper.RandString(6, "lower"))
	i, err := (&Locally{p}).Pin(b)
	if err != nil {
		t.Errorf("Unexpected pin data locally: %v", err)
	}
	if i != ipfsCid {
		t.Fatalf("Unexpected cid got %s instead of %s", i, ipfsCid)
	}
}
