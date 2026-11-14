// Package dashboard decides which Vite entries make up the admin page.
package dashboard

import (
	"fmt"
	"path"
	"path/filepath"
	"slices"

	"github.com/goravel/framework/contracts/modular"

	"goravel/modules/admin/contracts"
)

// Name is the admin module's name, Main its own entry, which mounts the admin.
const (
	Name = "admin"
	Main = "resources/js/main.ts"
)

// Entries returns the project-relative Vite entries of the admin page: every
// module's dashboard scripts in module order, then the admin's Main, so all
// modules have registered by the time the admin mounts.
func Entries(registry modular.Registry) ([]string, error) {
	var entries []string
	for _, module := range registry.All() {
		provider, ok := module.(contracts.ProvidesDashboardScripts)
		if !ok {
			continue
		}

		var declared []string
		if vite, ok := module.(modular.ProvidesViteEntries); ok {
			declared = vite.ViteEntries()
		}

		for _, script := range provider.DashboardScripts() {
			if !slices.Contains(declared, script) {
				return nil, fmt.Errorf("dashboard script [%s] of module [%s] is not in its ViteEntries()", script, module.Name())
			}

			entries = append(entries, entry(registry, module.Name(), script))
		}
	}

	if !registry.Has(Name) {
		return nil, fmt.Errorf("module [%s] is not registered", Name)
	}

	return append(entries, entry(registry, Name, Main)), nil
}

// entry joins a module-relative file onto the module directory, in the slash
// form Vite's manifest uses.
func entry(registry modular.Registry, name, file string) string {
	dir, _ := registry.Path(name)

	return path.Join(filepath.ToSlash(dir), file)
}
