package mcpserver

// EmitEvent bridges MCP-side events to the Wails frontend.
//
// It is injected by the host application (which owns the *application.App instance)
// at startup, because pkg/mcpserver cannot import the main package where wailsApp
// lives. When nil (e.g. when running under chying-cli without a Wails context),
// event emission is a no-op — MCP tools that emit events simply have no frontend
// to notify, which is safe.
var EmitEvent func(event string, data any)

// emit is a nil-safe helper. Returns false (and does nothing) when no emitter is
// wired, so callers can short-circuit or just fire-and-forget.
func emit(event string, data any) bool {
	if EmitEvent == nil {
		return false
	}
	EmitEvent(event, data)
	return true
}
