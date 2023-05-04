// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ipfs

import (
	"net"
	"net/http"
	"strconv"

	shell "github.com/ipfs/go-ipfs-api"
	pinner "github.com/wabarc/ipfs-pinner"
)

// PinningOption is a function type that modifies a Pinning struct by setting one of its fields.
// Each PinningOption function takes a pointer to a Pinning struct and returns nothing. When
// invoked, it modifies the Pinning struct by setting the appropriate field.
type PinningOption func(*Pinning)

// Mode sets the Mode field of a Pinning struct to the given mode.
func Mode(m mode) PinningOption {
	return func(o *Pinning) {
		o.Mode = m
	}
}

// Host sets the Host field of a Pinning struct to the given host.
func Host(h string) PinningOption {
	return func(o *Pinning) {
		o.Host = h
	}
}

// Port sets the Port field of a Pinning struct to the given port.
func Port(p int) PinningOption {
	return func(o *Pinning) {
		o.Port = p
	}
}

// Uses sets the Pinner field of a Pinning struct to the given name of the pinner.
func Uses(p string) PinningOption {
	return func(o *Pinning) {
		o.Pinner = p
	}
}

// Apikey sets the Apikey field of a Pinning struct to the given key.
func Apikey(k string) PinningOption {
	return func(o *Pinning) {
		o.Apikey = k
	}
}

// Secret sets the Secret field of a Pinning struct to the given key.
func Secret(s string) PinningOption {
	return func(o *Pinning) {
		o.Secret = s
	}
}

// Backoff sets the backoff field of a Pinning struct to the given boolean value.
func Backoff(b bool) PinningOption {
	return func(o *Pinning) {
		o.backoff = b
	}
}

// Client sets the Client field of a Pinning struct to the given http.Client instance.
func Client(c *http.Client) PinningOption {
	return func(o *Pinning) {
		o.Client = c
	}
}

// Options takes one or more PinningOptions and returns a Pinning struct has been configured
// according to those options.
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
