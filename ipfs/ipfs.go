// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ipfs

import (
	"bytes"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"

	shell "github.com/ipfs/go-ipfs-api"
	pinner "github.com/wabarc/ipfs-pinner"
)

type mode int

const (
	Remote mode = iota + 1 // Store files to pinning service
	Local                  // Store file on local IPFS node

	maxElapsedTime = time.Minute
	maxRetries     = 3
)

var _ Pinner = (*Locally)(nil)
var _ Pinner = (*Remotely)(nil)

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as IPFS handlers.
type HandlerFunc func(Pinner, interface{}) (string, error)

// Pinner is an interface that wraps the Pin method.
type Pinner interface {
	// Pin implements data transmission to the destination service by given buf. It
	// returns the content-id returned by the local IPFS server or a remote pinning service.
	Pin(buf []byte) (string, error)

	// Pin implements directory transmission to the destination service by given path. It
	// returns the content-id returned by the local IPFS server or a remote pinning service.
	PinDir(path string) (string, error)
}

// Locally embeds the Pinning struct, which provides configuration for pinning services
// used for data storage.
type Locally struct {
	Pinning
}

// Remotely embeds the Pinning struct, which provides configuration for pinning services
// used for data storage.
type Remotely struct {
	Pinning
}

// Pinning provides the configuration of pinning services that will be utilized for data storage.
type Pinning struct {
	// It supports both daemon server and remote pinner, defaults to remote pinner.
	Mode mode

	Host  string
	Port  int
	shell *shell.Shell // Only for daemon mode

	// For pinner mode, which normally requires the apikey and secret of the pinning service.
	Pinner string
	Apikey string
	Secret string

	// Client represents a http client.
	Client *http.Client

	// Whether or not to use backoff stragty.
	backoff bool
}

// Pin implements putting the data to local IPFS node by given buf. It
// returns content-id and an error.
func (l *Locally) Pin(buf []byte) (cid string, err error) {
	action := func() error {
		cid, err = l.shell.Add(bytes.NewReader(buf), shell.Pin(true))
		return err
	}
	err = l.doRetry(action)
	if err != nil {
		return "", errors.Wrap(err, "add file to IPFS failed")
	}
	return
}

// Pin implements putting the data to local IPFS node by given buf. It
// returns content-id and an error.
func (l *Locally) PinDir(path string) (cid string, err error) {
	action := func() error {
		cid, err = l.shell.AddDir(path)
		return err
	}
	err = l.doRetry(action)
	if err != nil {
		return "", errors.Wrap(err, "add directory to IPFS failed")
	}
	return
}

// Pin implements putting the data to destination pinning service by given buf. It
// returns content-id and an error.
func (r *Remotely) Pin(buf []byte) (cid string, err error) {
	action := func() error {
		cid, err = r.remotely().Pin(buf)
		return err
	}
	err = r.doRetry(action)
	return
}

// Pin implements putting the data to destination pinning service by given buf. It
// returns content-id and an error.
func (r *Remotely) PinDir(path string) (cid string, err error) {
	action := func() error {
		cid, err = r.remotely().Pin(path)
		return err
	}
	err = r.doRetry(action)
	return
}

func (r *Remotely) remotely() *pinner.Config {
	return &pinner.Config{
		Pinner: r.Pinner,
		Apikey: r.Apikey,
		Secret: r.Secret,
		Client: r.Client,
	}
}

func (p *Pinning) doRetry(op backoff.Operation) error {
	if p.backoff {
		exp := backoff.NewExponentialBackOff()
		exp.MaxElapsedTime = maxElapsedTime
		bo := backoff.WithMaxRetries(exp, maxRetries)

		return backoff.Retry(op, bo)
	}

	return op()
}
