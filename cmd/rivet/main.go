package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/wabarc/rivet"
	"github.com/wabarc/rivet/ipfs"

	pinner "github.com/wabarc/ipfs-pinner"
)

func main() {
	var (
		mode    string
		timeout uint
		// for local mode
		host string
		port int
		// for remote mode
		target string
		apikey string
		secret string
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage:\n\n")
		fmt.Fprintf(os.Stdout, "  rivet [options] [url1] ... [urlN]\n\n")

		flag.PrintDefaults()
	}
	basePrint := func() {
		fmt.Print("A toolkit makes it easier to archive webpages to IPFS.\n\n")
		flag.Usage()
		fmt.Fprint(os.Stdout, "\n")
	}

	flag.StringVar(&mode, "m", "remote", "Pin mode, supports mode: local, remote, archive")
	flag.UintVar(&timeout, "timeout", 30, "Timeout for every input URL")
	flag.StringVar(&host, "host", "localhost", "IPFS node address")
	flag.IntVar(&port, "port", 5001, "IPFS node port")
	flag.StringVar(&target, "t", "infura", "IPFS pinner, supports pinners: infura, pinata, nftstorage, web3storage.")
	flag.StringVar(&apikey, "u", "", "Pinner apikey or username.")
	flag.StringVar(&secret, "p", "", "Pinner sceret or password.")
	flag.Parse()

	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
	}
	if mode == "local" {
		opts = []ipfs.PinningOption{
			ipfs.Mode(ipfs.Local),
			ipfs.Host(host),
			ipfs.Port(port),
		}
	}

	switch target {
	case pinner.Infura, pinner.Pinata, pinner.NFTStorage, pinner.Web3Storage:
		opts = append(opts, ipfs.Uses(target), ipfs.Apikey(apikey), ipfs.Secret(secret))
	default:
		basePrint()
		fmt.Fprintln(os.Stderr, "Unknown target")
		os.Exit(0)
	}

	links := flag.Args()
	if len(links) < 1 {
		basePrint()
		fmt.Fprintln(os.Stderr, "link is missing")
		os.Exit(1)
	}

	ctx := context.Background()
	opt := ipfs.Options(opts...)
	toc := time.Duration(timeout) * time.Second
	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()

			input, err := url.Parse(link)
			if err != nil {
				fmt.Fprintf(os.Stderr, "rivet: %v\n", err)
				return
			}

			reqctx, cancel := context.WithTimeout(ctx, toc)
			defer cancel()

			r := &rivet.Shaft{Hold: opt, ArchiveOnly: mode == "archive"}
			if dest, err := r.Wayback(reqctx, input); err != nil {
				fmt.Fprintf(os.Stderr, "rivet: %v\n", err)
			} else {
				fmt.Fprintf(os.Stdout, "%s  %s\n", dest, link)
			}
		}(link)
	}
	wg.Wait()
}
