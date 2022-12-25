package rivet

import (
	"bufio"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/wabarc/helper"
	"github.com/wabarc/ipfs-pinner"
	"github.com/wabarc/rivet/ipfs"
)

var content = `<html>
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

func TestWayback(t *testing.T) {
	apikey := os.Getenv("IPFS_PINNER_PINATA_API_KEY")
	secret := os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
	if apikey == "" || secret == "" {
		t.Skip(`Must set env "IPFS_PINNER_PINATA_API_KEY" and "IPFS_PINNER_PINATA_SECRET_API_KEY"`)
	}

	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
		ipfs.Uses(pinner.Pinata),
		ipfs.Apikey(apikey),
		ipfs.Secret(secret),
	}
	opt := ipfs.Options(opts...)

	_, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	link := server.URL
	r := &Shaft{Hold: opt}
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
	apikey := os.Getenv("IPFS_PINNER_PINATA_API_KEY")
	secret := os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
	if apikey == "" || secret == "" {
		t.Skip(`Must set env "IPFS_PINNER_PINATA_API_KEY" and "IPFS_PINNER_PINATA_SECRET_API_KEY"`)
	}

	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
		ipfs.Uses(pinner.Pinata),
		ipfs.Apikey(apikey),
		ipfs.Secret(secret),
	}
	opt := ipfs.Options(opts...)

	_, mux, server := helper.MockServer()
	mux.HandleFunc("/", handleResponse)
	defer server.Close()

	link := server.URL
	r := &Shaft{Hold: opt}
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
