// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: Apache-2.0

package tscaddy

// directive.go contains the Tailscale directive for configuring node options at the virtual host level.

import (
	"net/http"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"tailscale.com/types/opt"
)

var (
	// siteConfigs stores site-specific node configurations
	// Key is the node name, value is the Node configuration
	siteConfigs   = make(map[string]Node)
	siteConfigsMu sync.RWMutex
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("tailscale", parseTailscaleDirective)
	httpcaddyfile.RegisterDirectiveOrder("tailscale", httpcaddyfile.After, "header")
}

// setSiteConfig stores a site-specific node configuration
func setSiteConfig(nodeName string, config Node) {
	siteConfigsMu.Lock()
	defer siteConfigsMu.Unlock()
	siteConfigs[nodeName] = config
}

// getSiteConfig retrieves a site-specific node configuration
func getSiteConfig(nodeName string) (Node, bool) {
	siteConfigsMu.RLock()
	defer siteConfigsMu.RUnlock()
	config, exists := siteConfigs[nodeName]
	return config, exists
}

// TailscaleDirective is a Caddy HTTP handler that configures Tailscale node options
// for the current virtual host. This allows overriding global Tailscale configuration
// on a per-site basis.
type TailscaleDirective struct {
	// NodeName is the name of the Tailscale node to configure.
	// If empty, it will be derived from the bind address.
	NodeName string `json:"node_name,omitempty"`

	// AuthKey is the Tailscale auth key used to register the node.
	AuthKey string `json:"auth_key,omitempty"`

	// ControlURL specifies the control URL to use for the node.
	ControlURL string `json:"control_url,omitempty"`

	// Ephemeral specifies whether the node should be registered as ephemeral.
	Ephemeral opt.Bool `json:"ephemeral,omitempty"`

	// WebUI specifies whether the node should run the Web UI for remote management.
	WebUI opt.Bool `json:"webui,omitempty"`

	// Hostname is the hostname to use when registering the node.
	Hostname string `json:"hostname,omitempty"`

	// Port is the port to use for the Tailscale node.
	Port uint16 `json:"port,omitempty"`

	// StateDir specifies the state directory for the node.
	StateDir string `json:"state_dir,omitempty"`

	// Tags specifies the list of tags to apply to this node.
	Tags []string `json:"tags,omitempty"`
}

func (TailscaleDirective) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.tailscale",
		New: func() caddy.Module { return new(TailscaleDirective) },
	}
}

// Provision implements caddy.Provisioner.
func (t *TailscaleDirective) Provision(ctx caddy.Context) error {
	// Use the node name that was set during parsing
	nodeName := t.NodeName
	if nodeName == "" {
		nodeName = "default"
	}

	// Create a Node configuration from the directive settings
	node := Node{
		AuthKey:    t.AuthKey,
		ControlURL: t.ControlURL,
		Ephemeral:  t.Ephemeral,
		WebUI:      t.WebUI,
		Hostname:   t.Hostname,
		Port:       t.Port,
		StateDir:   t.StateDir,
		Tags:       t.Tags,
		name:       nodeName,
	}

	// Store the configuration globally so it can be accessed during node creation
	setSiteConfig(nodeName, node)

	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
// This directive doesn't actually handle HTTP requests - it just configures the Tailscale node.
// So we pass through to the next handler.
func (t TailscaleDirective) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return next.ServeHTTP(w, r)
}

// parseTailscaleDirective parses the tailscale directive from a Caddyfile.
func parseTailscaleDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var directive TailscaleDirective

	// Parse the directive arguments
	for h.Next() {
		// First argument could be the node name
		if h.NextArg() {
			directive.NodeName = h.Val()
		}

		// If no node name was provided as an argument, use "default"
		// Users can explicitly specify a node name if they want site-specific config
		// that differs from the global configuration
		if directive.NodeName == "" {
			directive.NodeName = "default"
		}

		// Create a temporary Node to use with the shared parsing function
		node := Node{}
		err := parseNodeOptionsFromHelper(h, &node)
		if err != nil {
			return nil, err
		}

		// Copy the parsed values to the directive
		directive.AuthKey = node.AuthKey
		directive.ControlURL = node.ControlURL
		directive.Ephemeral = node.Ephemeral
		directive.WebUI = node.WebUI
		directive.Hostname = node.Hostname
		directive.Port = node.Port
		directive.StateDir = node.StateDir
		directive.Tags = node.Tags
	}

	return directive, nil
}

var (
	_ caddy.Provisioner           = (*TailscaleDirective)(nil)
	_ caddyhttp.MiddlewareHandler = (*TailscaleDirective)(nil)
)
