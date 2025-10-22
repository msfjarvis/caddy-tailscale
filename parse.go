// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: Apache-2.0

package tscaddy

// parse.go contains shared parsing functions for Tailscale configuration

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"tailscale.com/types/opt"
)

// parseNodeOptionsFromDispenser parses common node configuration options from a caddyfile.Dispenser.
func parseNodeOptionsFromDispenser(d *caddyfile.Dispenser, node *Node) error {
	for d.NextBlock(0) {
		switch d.Val() {
		case "auth_key":
			if !d.NextArg() {
				return d.ArgErr()
			}
			node.AuthKey = d.Val()

		case "control_url":
			if !d.NextArg() {
				return d.ArgErr()
			}
			node.ControlURL = d.Val()

		case "ephemeral":
			if d.NextArg() {
				v, err := strconv.ParseBool(d.Val())
				if err != nil {
					return d.WrapErr(err)
				}
				node.Ephemeral = opt.NewBool(v)
			} else {
				node.Ephemeral = opt.NewBool(true)
			}

		case "hostname":
			if !d.NextArg() {
				return d.ArgErr()
			}
			node.Hostname = d.Val()

		case "port":
			if !d.NextArg() {
				return d.ArgErr()
			}
			v, err := strconv.ParseUint(d.Val(), 10, 16)
			if err != nil {
				return d.WrapErr(err)
			}
			node.Port = uint16(v)

		case "state_dir":
			if !d.NextArg() {
				return d.ArgErr()
			}
			node.StateDir = d.Val()

		case "webui":
			if d.NextArg() {
				v, err := strconv.ParseBool(d.Val())
				if err != nil {
					return d.WrapErr(err)
				}
				node.WebUI = opt.NewBool(v)
			} else {
				node.WebUI = opt.NewBool(true)
			}

		case "tags":
			for d.NextArg() {
				node.Tags = append(node.Tags, d.Val())
			}

		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}

// parseNodeOptionsFromHelper parses common node configuration options from an httpcaddyfile.Helper.
func parseNodeOptionsFromHelper(h interface {
	NextBlock(int) bool
	Val() string
	NextArg() bool
	ArgErr() error
	WrapErr(error) error
	Errf(string, ...interface{}) error
}, node *Node) error {
	for h.NextBlock(0) {
		switch h.Val() {
		case "auth_key":
			if !h.NextArg() {
				return h.ArgErr()
			}
			node.AuthKey = h.Val()

		case "control_url":
			if !h.NextArg() {
				return h.ArgErr()
			}
			node.ControlURL = h.Val()

		case "ephemeral":
			if h.NextArg() {
				v, err := strconv.ParseBool(h.Val())
				if err != nil {
					return h.WrapErr(err)
				}
				node.Ephemeral = opt.NewBool(v)
			} else {
				node.Ephemeral = opt.NewBool(true)
			}

		case "hostname":
			if !h.NextArg() {
				return h.ArgErr()
			}
			node.Hostname = h.Val()

		case "port":
			if !h.NextArg() {
				return h.ArgErr()
			}
			v, err := strconv.ParseUint(h.Val(), 10, 16)
			if err != nil {
				return h.WrapErr(err)
			}
			node.Port = uint16(v)

		case "state_dir":
			if !h.NextArg() {
				return h.ArgErr()
			}
			node.StateDir = h.Val()

		case "webui":
			if h.NextArg() {
				v, err := strconv.ParseBool(h.Val())
				if err != nil {
					return h.WrapErr(err)
				}
				node.WebUI = opt.NewBool(v)
			} else {
				node.WebUI = opt.NewBool(true)
			}

		case "tags":
			for h.NextArg() {
				node.Tags = append(node.Tags, h.Val())
			}

		default:
			return h.Errf("unrecognized subdirective: %s", h.Val())
		}
	}
	return nil
}

// parseAppOptions parses app-level configuration options from a caddyfile.Dispenser.
// This function handles options that are specific to the global app configuration.
func parseAppOptions(d *caddyfile.Dispenser, app *App) error {
	for d.NextBlock(0) {
		switch d.Val() {
		case "auth_key":
			if !d.NextArg() {
				return d.ArgErr()
			}
			app.DefaultAuthKey = d.Val()

		case "control_url":
			if !d.NextArg() {
				return d.ArgErr()
			}
			app.ControlURL = d.Val()

		case "ephemeral":
			if d.NextArg() {
				v, err := strconv.ParseBool(d.Val())
				if err != nil {
					return d.WrapErr(err)
				}
				app.Ephemeral = v
			} else {
				app.Ephemeral = true
			}

		case "state_dir":
			if !d.NextArg() {
				return d.ArgErr()
			}
			app.StateDir = d.Val()

		case "webui":
			if d.NextArg() {
				v, err := strconv.ParseBool(d.Val())
				if err != nil {
					return d.WrapErr(err)
				}
				app.WebUI = v
			} else {
				app.WebUI = true
			}

		case "tags":
			for d.NextArg() {
				app.Tags = append(app.Tags, d.Val())
			}

		default:
			// Try to parse as a named node configuration
			node, err := parseNamedNodeConfig(d)
			if err != nil {
				return err
			}
			if app.Nodes == nil {
				app.Nodes = make(map[string]Node)
			}
			app.Nodes[node.name] = node
		}
	}
	return nil
}

// parseNamedNodeConfig parses a named node configuration block.
// This is used for parsing node configurations within the global app config.
func parseNamedNodeConfig(d *caddyfile.Dispenser) (Node, error) {
	name := d.Val()
	segment := d.NewFromNextSegment()

	if !segment.Next() {
		return Node{}, d.ArgErr()
	}

	node := Node{name: name}
	err := parseNodeOptionsFromDispenser(segment, &node)
	if err != nil {
		return node, err
	}

	return node, nil
}
