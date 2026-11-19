package feature

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/module"
	frameworkmodular "github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/path"
	goravelvite "github.com/goravel/framework/vite"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/modules/admin"
	"goravel/modules/admin/dashboard"
	"goravel/modules/blog"
	"goravel/tests"
)

// adminBuildDirectory is under public/build, which is gitignored whole.
const adminBuildDirectory = "build/admin-test"

// adminManifest is what `vite build` writes for the two dashboard entries: both
// import a shared chunk, which the page preloads once.
const adminManifest = `{
  "modules/blog/resources/js/dashboard.ts": {"file": "assets/dashboard-b1.js", "src": "modules/blog/resources/js/dashboard.ts", "isEntry": true, "imports": ["_shared-s1.js"]},
  "modules/admin/resources/js/main.ts": {"file": "assets/main-a1.js", "src": "modules/admin/resources/js/main.ts", "isEntry": true, "css": ["assets/main-a1.css"], "imports": ["_shared-s1.js"]},
  "_shared-s1.js": {"file": "assets/shared-s1.js"}
}`

// AdminDashboardTestSuite covers the admin module's own capability,
// contracts.ProvidesDashboardScripts: modules list Vite entries, the admin
// page loads them before its main.ts, and each script appends its widget.
type AdminDashboardTestSuite struct {
	suite.Suite
	tests.TestCase
	originalBuildDirectory string
	originalHotFile        string
}

func TestAdminDashboardTestSuite(t *testing.T) {
	suite.Run(t, &AdminDashboardTestSuite{})
}

func (s *AdminDashboardTestSuite) SetupSuite() {
	s.originalBuildDirectory = facades.Vite().BuildDirectory()
	s.originalHotFile = facades.Config().GetString("vite.hot_file", goravelvite.DefaultHotFile())
}

func (s *AdminDashboardTestSuite) TearDownTest() {
	vite, ok := facades.Vite().(*goravelvite.Vite)
	s.Require().True(ok)
	vite.UseBuildDirectory(s.originalBuildDirectory).UseHotFile(s.originalHotFile)

	s.Require().NoError(os.RemoveAll(path.Public(adminBuildDirectory)))
}

// useFixtureVite points the shared Vite at an empty fixture build directory.
func (s *AdminDashboardTestSuite) useFixtureVite() {
	vite, ok := facades.Vite().(*goravelvite.Vite)
	s.Require().True(ok)
	vite.UseBuildDirectory(adminBuildDirectory).UseHotFile(adminBuildDirectory + "/hot")
}

// The fakes live in this package, so their directory is "tests/feature".
type dashboardModule struct {
	name         string
	dependencies []string
	entries      []string
	scripts      []string
}

func (m *dashboardModule) Name() string               { return m.name }
func (m *dashboardModule) Dependencies() []string     { return m.dependencies }
func (m *dashboardModule) ViteEntries() []string      { return m.entries }
func (m *dashboardModule) DashboardScripts() []string { return m.scripts }

// plainModule has no dashboard scripts and is skipped.
type plainModule struct{}

func (m *plainModule) Name() string { return "plain" }

func (s *AdminDashboardTestSuite) TestEntries() {
	tests := []struct {
		name        string
		modules     []module.Module
		expect      []string
		expectError error
	}{
		{
			name:    "should load the blog's script before the admin's main",
			modules: []module.Module{&admin.Module{}, &blog.Module{}},
			expect:  []string{"modules/blog/resources/js/dashboard.ts", "modules/admin/resources/js/main.ts"},
		},
		{
			name: "should follow dependency order and skip modules without dashboard scripts",
			modules: []module.Module{
				&admin.Module{},
				&dashboardModule{name: "reports", dependencies: []string{"shop"}, entries: []string{"resources/js/reports.ts"}, scripts: []string{"resources/js/reports.ts"}},
				&plainModule{},
				&dashboardModule{name: "shop", entries: []string{"resources/js/a.ts", "resources/js/b.ts"}, scripts: []string{"resources/js/a.ts", "resources/js/b.ts"}},
			},
			expect: []string{
				"tests/feature/resources/js/a.ts",
				"tests/feature/resources/js/b.ts",
				"tests/feature/resources/js/reports.ts",
				"modules/admin/resources/js/main.ts",
			},
		},
		{
			name: "should refuse a script that is not a Vite entry, since nothing would build it",
			modules: []module.Module{
				&admin.Module{},
				&dashboardModule{name: "shop", scripts: []string{"resources/js/shop.ts"}},
			},
			expectError: errors.New("dashboard script [resources/js/shop.ts] of module [shop] is not in its ViteEntries()"),
		},
		{
			name:        "should need the admin module",
			modules:     []module.Module{&blog.Module{}},
			expectError: errors.New("module [admin] is not registered"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			registry := frameworkmodular.NewRegistry("")
			s.Require().NoError(registry.Register(tt.modules...))

			entries, err := dashboard.Entries(registry)
			s.Equal(tt.expect, entries)
			s.Equal(tt.expectError, err)
		})
	}
}

// TestPage renders the admin from a built manifest: one Vite call for every
// entry, the shared chunk preloaded once, the blog's script before main.ts.
func (s *AdminDashboardTestSuite) TestPage() {
	s.useFixtureVite()

	manifest := path.Public(adminBuildDirectory, ".vite", "manifest.json")
	s.Require().NoError(os.MkdirAll(filepath.Dir(manifest), 0o755))
	s.Require().NoError(os.WriteFile(manifest, []byte(adminManifest), 0o644))

	resp, err := s.Http(s.T()).Get("/admin")
	s.Require().NoError(err)
	resp.AssertOk().AssertSeeInOrder([]string{
		"<title>Admin</title>",
		`<link rel="modulepreload" href="/build/admin-test/assets/shared-s1.js">`,
		`<link rel="stylesheet" href="/build/admin-test/assets/main-a1.css">`,
		`<script type="module" src="/build/admin-test/assets/dashboard-b1.js"></script>`,
		`<script type="module" src="/build/admin-test/assets/main-a1.js"></script>`,
		`<main id="widgets" class="admin-widgets"></main>`,
	}, false)
}

// TestPage_WithoutBuild reports the Vite error instead of a page without scripts.
func (s *AdminDashboardTestSuite) TestPage_WithoutBuild() {
	s.useFixtureVite()

	resp, err := s.Http(s.T()).Get("/admin")
	s.Require().NoError(err)
	resp.AssertInternalServerError()
}
