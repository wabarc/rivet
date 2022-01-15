# Rivet

[![LICENSE](https://img.shields.io/github/license/wabarc/rivet.svg?color=green)](https://github.com/wabarc/rivet/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wabarc/rivet)](https://goreportcard.com/report/github.com/wabarc/rivet)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/wabarc/rivet/Go?color=brightgreen)](https://github.com/wabarc/rivet/actions)
[![Go Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/wabarc/rivet)
[![Releases](https://img.shields.io/github/v/release/wabarc/rivet.svg?include_prereleases&color=blue)](https://github.com/wabarc/rivet/releases)

Rivet is both a command-line tool and a Golang package for archiving webpages to IPFS.

Supported Golang version: See [.github/workflows/testing.yml](./.github/workflows/testing.yml)

## Installation

The simplest, cross-platform way is to download from [GitHub Releases](https://github.com/wabarc/rivet/releases) and place the executable file in your PATH.

From source:

```sh
go get -u github.com/wabarc/rivet/cmd/rivet
```

From [GoBinaries](https://gobinaries.com/):

```sh
curl -sf https://gobinaries.com/wabarc/rivet/cmd/rivet | sh
```

Using [Snapcraft](https://snapcraft.io/rivet) (on GNU/Linux)

```sh
sudo snap install rivet
```

## Usage

### Command line

```sh
A toolkit makes it easier to archive webpages to IPFS.

Usage:

  rivet [options] [url1] ... [urlN]

  -h string
        IPFS node address (default "localhost")
  -k string
        Pinner apikey or username.
  -m string
        Pin mode, supports mode: local, remote (default "remote")
  -p int
        IPFS node port (default 5001)
  -s string
        Pinner sceret or password.
  -t string
        IPFS pinner, supports pinners: infura, pinata, nftstorage, web3storage. (default "infura")
  -timeout uint
        Timeout for every input URL (default 30)
```

#### Examples

Stores data on local IPFS node.

```sh
rivet -m local https://example.com https://example.org
```

Stores data to remote pinning services.

```sh
rivet https://example.com
```

Or, specify a pinning service.

```sh
rivet -t pinata -k your-apikey -s your-secret https://example.com
```

### Go package

<!-- markdownlint-disable MD010 -->
```go
package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/wabarc/ipfs-pinner"
	"github.com/wabarc/rivet"
	"github.com/wabarc/rivet/ipfs"
)

func main() {
	opts := []ipfs.PinningOption{
		ipfs.Mode(ipfs.Remote),
		ipfs.Uses(pinner.Infura),
	}
	p := ipfs.Options(opts...)
	r := &rivet.Shaft{Pinning: p}
	l := "https://example.com"
	input, err := url.Parse(l)
	if err != nil {
		panic(err)
	}

	dst, err := r.Wayback(context.TODO(), input)
	if err != nil {
		panic(err)
	}
	fmt.Println(dst)
}
```
<!-- markdownlint-enable MD010 -->

## F.A.Q

### Does not load resources after redirecting to a subdomain?

Subdomain gateway such as https://baf**.ipfs.io/ are not yet supported.

See: https://docs.ipfs.io/how-to/address-ipfs-on-web/#subdomain-gateway

### Optional to disable JavaScript for some URI?

If you prefer to disable JavaScript on saving webpage, you could add environmental variables `DISABLEJS_URIS`
and set the values with the following formats:

```sh
export DISABLEJS_URIS=wikipedia.org|eff.org/tags
```

It will disable JavaScript for domain of the `wikipedia.org` and path of the `eff.org/tags` if matching it.

## Credit

Special thanks to [@RadhiFadlillah](https://github.com/RadhiFadlillah) for making [obelisk](https://github.com/go-shiori/obelisk), under which the crawling of the web is based.

## Contributing

We encourage all contributions to this repository! Open an issue! Or open a Pull Request!

## License

This software is released under the terms of the MIT. See the [LICENSE](https://github.com/wabarc/rivet/blob/main/LICENSE) file for details.

