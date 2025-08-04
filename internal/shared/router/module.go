package router

import (
	"go.uber.org/fx"
)

// Module provides the router configuration
var Module = fx.Module("router",
	// Router setup is now handled in main.go's NewHTTPServer
	// This module exists to maintain architectural consistency
	// and could be extended with router-specific providers in the future
)