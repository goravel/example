// Package contracts holds the capabilities other modules implement to take part
// in the admin. Go interfaces are structural, so implementing one needs no import.
package contracts

// ProvidesDashboardScripts lists the Vite entries the admin loads before it
// mounts, relative to the module directory, e.g. "resources/js/dashboard.ts".
// Each one must also be in the module's ViteEntries(), which is what builds it.
type ProvidesDashboardScripts interface {
	DashboardScripts() []string
}
