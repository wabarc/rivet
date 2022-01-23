// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ipfs

import (
	"bytes"
	"net"
	"strconv"

	"github.com/pkg/errors"

	shell "github.com/ipfs/go-ipfs-api"
	pinner "github.com/wabarc/ipfs-pinner"
)

type mode int

const (
	Remote mode = iota // Store files to pinning service
	Local              // Store file on local IPFS node
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

type Locally struct {
	Pinning
}

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
}

type PinningOption func(*Pinning)

func Mode(m mode) PinningOption {
	return func(o *Pinning) {
		o.Mode = m
	}
}

func Host(h string) PinningOption {
	return func(o *Pinning) {
		o.Host = h
	}
}

func Port(p int) PinningOption {
	return func(o *Pinning) {
		o.Port = p
	}
}

func Uses(p string) PinningOption {
	return func(o *Pinning) {
		o.Pinner = p
	}
}

func Apikey(k string) PinningOption {
	return func(o *Pinning) {
		o.Apikey = k
	}
}

func Secret(s string) PinningOption {
	return func(o *Pinning) {
		o.Secret = s
	}
}

func Options(options ...PinningOption) Pinning {
	var p Pinning
	for _, o := range options {
		o(&p)
	}
	if p.Mode == Local {
		p.shell = shell.NewShell(net.JoinHostPort(p.Host, strconv.Itoa(p.Port)))
	}
	if p.Mode == Remote && p.Pinner == "" {
		p.Pinner = pinner.Infura
	}
	return p
}

// Pin implements putting the data to local IPFS node by given buf. It
// returns content-id and an error.
func (l *Locally) Pin(buf []byte) (cid string, err error) {
	cid, err = l.shell.Add(bytes.NewReader(buf), shell.Pin(true))
	if err != nil {
		return "", errors.Wrap(err, "add file to IPFS failed")
	}
	return
}

// Pin implements putting the data to local IPFS node by given buf. It
// returns content-id and an error.
func (l *Locally) PinDir(path string) (cid string, err error) {
	cid, err = l.shell.AddDir(path)
	if err != nil {
		return "", errors.Wrap(err, "add directory to IPFS failed")
	}
	return
}

// Pin implements putting the data to destination pinning service by given buf. It
// returns content-id and an error.
func (r *Remotely) Pin(buf []byte) (cid string, err error) {
	pinner := &pinner.Config{
		Pinner: r.Pinner,
		Apikey: r.Apikey,
		Secret: r.Secret,
	}
	cid, err = pinner.Pin(buf)
	return
}

// Pin implements putting the data to destination pinning service by given buf. It
// returns content-id and an error.
func (r *Remotely) PinDir(path string) (cid string, err error) {
	pinner := &pinner.Config{
		Pinner: r.Pinner,
		Apikey: r.Apikey,
		Secret: r.Secret,
	}
	cid, err = pinner.Pin(path)
	return
}
