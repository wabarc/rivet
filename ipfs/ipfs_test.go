package ipfs

import (
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner"
)

var (
	apikey           = "1234"
	secret           = "abcd"
	badRequestJSON   = `{}`
	unauthorizedJSON = `{}`
	pinHashJSON      = `{
    "hashToPin": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
}`
	pinFileJSON = `{
    "IpfsHash": "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a",
    "PinSize": 1234,
    "Timestamp": "1979-01-01 00:00:00Z"
}`
	ipfsCid = "Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"
	addJSON = fmt.Sprintf(`{
  "Bytes": 0,
  "Hash": "%s",
  "Name": "name",
  "Size": "1B"
}`, ipfsCid)
)

func handleResponse(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Hostname() {
	case "api.pinata.cloud":
		authorization := r.Header.Get("Authorization")
		apiKey := r.Header.Get("pinata_api_key")
		apiSec := r.Header.Get("pinata_secret_api_key")
		switch {
		case apiKey != "" && apiSec != "":
			// access
		case authorization != "" && !strings.HasPrefix(authorization, "Bearer"):
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(unauthorizedJSON))
			return
		default:
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(unauthorizedJSON))
			return
		}

		switch r.URL.Path {
		case "/pinning/pinFileToIPFS":
			_ = r.ParseMultipartForm(32 << 20)
			_, params, parseErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if parseErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(badRequestJSON))
				return
			}

			multipartReader := multipart.NewReader(r.Body, params["boundary"])
			defer r.Body.Close()

			// Pin directory
			if multipartReader != nil && len(r.MultipartForm.File["file"]) > 1 {
				_, _ = w.Write([]byte(pinFileJSON))
				return
			}
			// Pin file
			if multipartReader != nil && len(r.MultipartForm.File["file"]) == 1 {
				_, _ = w.Write([]byte(pinFileJSON))
				return
			}
		case "/pinning/pinByHash":
			_, _ = w.Write([]byte(pinHashJSON))
			return
		}
	}
}

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
	client, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	opts := []PinningOption{
		Mode(Remote),
		Uses(pinner.Pinata),
		Apikey(apikey),
		Secret(secret),
		Client(client),
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
