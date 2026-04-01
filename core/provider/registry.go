package provider

import (
	"github.com/my-paas/core/provider/golang"
	"github.com/my-paas/core/provider/java"
	"github.com/my-paas/core/provider/node"
	"github.com/my-paas/core/provider/php"
	"github.com/my-paas/core/provider/python"
	"github.com/my-paas/core/provider/rust"
	"github.com/my-paas/core/provider/staticfile"
)

// GetProviders returns all providers in priority order.
// First match wins (same pattern as Railpack).
func GetProviders() []Provider {
	return []Provider{
		&php.PhpProvider{},
		&golang.GoProvider{},
		&java.JavaProvider{},
		&rust.RustProvider{},
		&python.PythonProvider{},
		&node.NodeProvider{},
		&staticfile.StaticfileProvider{},
	}
}
