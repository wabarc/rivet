// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package rivet

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/pkg/errors"
	"github.com/wabarc/rivet/internal/obelisk"
	"github.com/wabarc/rivet/ipfs"
)

// Shaft represents the rivet handler.
type Shaft struct {
	// Hold specifies which IPFS mode to pin data through.
	Hold ipfs.Pinning

	ArchiveOnly bool // Do not store file on any IPFS node, just archive
}

// Wayback uses IPFS to archive webpages.
func (s *Shaft) Wayback(ctx context.Context, input *url.URL) (cid string, err error) {
	name := sanitize.BaseName(input.Host) + sanitize.BaseName(input.Path)
	dir := "rivet-" + name
	if len(dir) > 255 {
		dir = dir[:254]
	}

	dir, err = ioutil.TempDir(os.TempDir(), dir+"-")
	if err != nil {
		return "", errors.Wrap(err, "create temp directory failed: "+dir)
	}
	defer os.RemoveAll(dir)

	uri := input.String()
	req := obelisk.Request{URL: uri, Input: inputFromContext(ctx)}
	// http.DefaultTransport
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	arc := &obelisk.Archiver{
		DisableJS: isDisableJS(uri),

		SkipResourceURLError: true,

		ResTempDir: dir,

		SingleFile: s.ArchiveOnly,

		Transport: tr,
	}
	arc.Validate()

	content, _, err := arc.Archive(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "archive failed")
	}

	// For auto indexing in IPFS, the filename should be index.html.
	indexFile := filepath.Join(dir, "index.html")
	if s.ArchiveOnly {
		indexFile = name + ".html"
	}

	if err := ioutil.WriteFile(indexFile, content, 0600); err != nil {
		return "", errors.Wrap(err, "create index file failed")
	}

	if s.ArchiveOnly {
		return indexFile, nil
	}

	switch s.Hold.Mode {
	case ipfs.Local:
		cid, err = (&ipfs.Locally{Pinning: s.Hold}).PinDir(dir)
	case ipfs.Remote:
		cid, err = (&ipfs.Remotely{Pinning: s.Hold}).PinDir(dir)
	}
	if err != nil {
		return "", errors.Wrap(err, "pin failed")
	}
	if cid == "" {
		return "", errors.New("cid empty")
	}

	return "https://ipfs.io/ipfs/" + cid, nil
}

type ctxKeyInput struct{}

// WithInput permits to inject a webpage into a context by given input.
func (s *Shaft) WithInput(ctx context.Context, input []byte) (c context.Context) {
	return context.WithValue(ctx, ctxKeyInput{}, input)
}

func inputFromContext(ctx context.Context) io.Reader {
	if b, ok := ctx.Value(ctxKeyInput{}).([]byte); ok {
		return bytes.NewReader(b)
	}
	return nil
}

func isDisableJS(link string) bool {
	// e.g. DISABLEJS_URIS=wikipedia.org|eff.org/tags
	uris := os.Getenv("DISABLEJS_URIS")
	if uris == "" {
		return false
	}

	regex := regexp.QuoteMeta(strings.ReplaceAll(uris, "|", "@@"))
	re := regexp.MustCompile(`(?m)` + strings.ReplaceAll(regex, "@@", "|"))

	return re.MatchString(link)
}
