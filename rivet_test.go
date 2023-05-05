package rivet

import (
	"bufio"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner"
	"github.com/wabarc/rivet/ipfs"
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
	content = `<html>
<head>
    <title>Example Domain</title>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is for use in illustrative examples in documents. You may use this
    domain in literature without prior coordination or asking for permission.</p>
    <p><a href="https://www.iana.org/domains/example">More information...</a></p>
    <p><img src="/image.png"></p>
</div>
</body>
</html>
`
)

func genImage(height int) bytes.Buffer {
	width := 1024

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	var b bytes.Buffer
	f := bufio.NewWriter(&b)
	png.Encode(f, img) // Encode as PNG.

	return b
}

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
	default:
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(content))
		case "/image.png":
			buf := genImage(1024)
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(buf.Bytes())
		}
	}
}

func TestWayback(t *testing.T) {
	client, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
		ipfs.Uses(pinner.Pinata),
		ipfs.Apikey(apikey),
		ipfs.Secret(secret),
		ipfs.Client(client),
	}
	opt := ipfs.Options(opts...)

	link := server.URL
	r := &Shaft{Hold: opt, Client: client}
	input, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.Wayback(context.TODO(), input)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaybackWithInput(t *testing.T) {
	client, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
		ipfs.Uses(pinner.Pinata),
		ipfs.Apikey(apikey),
		ipfs.Secret(secret),
		ipfs.Client(client),
	}
	opt := ipfs.Options(opts...)

	link := server.URL
	r := &Shaft{Hold: opt, Client: client}
	input, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}

	ctx := r.WithInput(context.TODO(), []byte(content))
	_, err = r.Wayback(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
}
