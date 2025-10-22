// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: Apache-2.0

package tscaddy

// app.go contains App and Node, which provide global configuration for registering Tailscale nodes.

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"go.uber.org/zap"
	"tailscale.com/types/opt"
)

func init() {
	caddy.RegisterModule(App{})
	httpcaddyfile.RegisterGlobalOption("tailscale", parseAppConfig)
}

// App is the Tailscale Caddy app used to configure Tailscale nodes.
// Nodes can be used to serve sites privately on a Tailscale network,
// or to connect to other Tailnet nodes as upstream proxy backend.
type App struct {
	// DefaultAuthKey is the default auth key to use for Tailscale if no other auth key is specified.
	DefaultAuthKey string `json:"auth_key,omitempty" caddy:"namespace=tailscale.auth_key"`

	// ControlURL specifies the default control URL to use for nodes.
	ControlURL string `json:"control_url,omitempty" caddy:"namespace=tailscale.control_url"`

	// Ephemeral specifies whether Tailscale nodes should be registered as ephemeral.
	Ephemeral bool `json:"ephemeral,omitempty" caddy:"namespace=tailscale.ephemeral"`

	// StateDir specifies the default state directory for Tailscale nodes.
	// Each node will have a subdirectory under this parent directory for its state.
	StateDir string `json:"state_dir,omitempty" caddy:"namespace=tailscale.state_dir"`

	// WebUI specifies whether Tailscale nodes should run the Web UI for remote management.
	WebUI bool `json:"webui,omitempty" caddy:"namespace=tailscale.webui"`

	// Tags specifies the list of tags to apply to all nodes.
	Tags []string `json:"tags,omitempty" caddy:"namespace=tailscale.tags"`

	// Nodes is a map of per-node configuration which overrides global options.
	Nodes map[string]Node `json:"nodes,omitempty" caddy:"namespace=tailscale"`

	logger *zap.Logger
}

// Node is a Tailscale node configuration.
// A single node can be used to serve multiple sites on different domains or ports,
// and/or to connect to other Tailscale nodes.
type Node struct {
	// AuthKey is the Tailscale auth key used to register the node.
	AuthKey string `json:"auth_key,omitempty" caddy:"namespace=auth_key"`

	// ControlURL specifies the control URL to use for the node.
	ControlURL string `json:"control_url,omitempty" caddy:"namespace=tailscale.control_url"`

	// Ephemeral specifies whether the node should be registered as ephemeral.
	Ephemeral opt.Bool `json:"ephemeral,omitempty" caddy:"namespace=tailscale.ephemeral"`

	// WebUI specifies whether the node should run the Web UI for remote management.
	WebUI opt.Bool `json:"webui,omitempty" caddy:"namespace=tailscale.webui"`

	// Hostname is the hostname to use when registering the node.
	Hostname string `json:"hostname,omitempty" caddy:"namespace=tailscale.hostname"`

	Port uint16 `json:"port,omitempty" caddy:"namespace=tailscale.port"`

	// StateDir specifies the state directory for the node.
	StateDir string `json:"state_dir,omitempty" caddy:"namespace=tailscale.state_dir"`

	// Tags specifies the list of tags to apply to this node.
	Tags []string `json:"tags,omitempty" caddy:"namespace=tailscale.tags"`

	name string
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "tailscale",
		New: func() caddy.Module { return new(App) },
	}
}

func (t *App) Provision(ctx caddy.Context) error {
	t.logger = ctx.Logger(t)
	return nil
}

func (t *App) Start() error {
	return nil
}

func (t *App) Stop() error {
	return nil
}

func parseAppConfig(d *caddyfile.Dispenser, _ any) (any, error) {
	app := &App{
		Nodes: make(map[string]Node),
	}
	if !d.Next() {
		return app, d.ArgErr()
	}

	err := parseAppOptions(d, app)
	if err != nil {
		return nil, err
	}

	return httpcaddyfile.App{
		Name:  "tailscale",
		Value: caddyconfig.JSON(app, nil),
	}, nil
}

func parseNodeConfig(d *caddyfile.Dispenser) (Node, error) {
	return parseNamedNodeConfig(d)
}

var (
	_ caddy.App         = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
)
