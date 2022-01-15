// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package rivet

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/wabarc/rivet/internal/obelisk"
	"github.com/wabarc/rivet/ipfs"
)

// Shaft represents the rivet handler.
type Shaft struct {
	// Hold specifies which IPFS mode to pin data through.
	Hold ipfs.Pinning
}

// Wayback uses IPFS to archive webpages.
func (s *Shaft) Wayback(ctx context.Context, input *url.URL) (cid string, err error) {
	dir, err := ioutil.TempDir(os.TempDir(), "rivet-")
	if err != nil {
		return "", errors.Wrap(err, "create temp directory failed")
	}
	defer os.RemoveAll(dir)

	var pinFunc ipfs.HandlerFunc
	switch s.Hold.Mode {
	case ipfs.Local:
		pinFunc = func(i ipfs.Pinner, b []byte) (string, error) {
			return (&ipfs.Locally{Pinning: s.Hold}).Pin(b)
		}
	case ipfs.Remote:
		pinFunc = func(i ipfs.Pinner, b []byte) (string, error) {
			return (&ipfs.Remotely{Pinning: s.Hold}).Pin(b)
		}
	default:
		return "", errors.New("unknown pinning mode")
	}

	uri := input.String()
	req := obelisk.Request{URL: uri, Input: inputFromContext(ctx)}
	arc := &obelisk.Archiver{
		DisableJS: isDisableJS(uri),

		SkipResourceURLError: true,

		PinFunc: pinFunc,
	}
	arc.Validate()

	content, _, err := arc.Archive(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "archive failed")
	}

	switch s.Hold.Mode {
	case ipfs.Local:
		cid, err = (&ipfs.Locally{Pinning: s.Hold}).Pin(content)
	case ipfs.Remote:
		cid, err = (&ipfs.Remotely{Pinning: s.Hold}).Pin(content)
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
