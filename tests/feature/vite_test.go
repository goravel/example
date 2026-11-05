package feature

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/path"
	goravelvite "github.com/goravel/framework/vite"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// fixtureBuildDirectory keeps the fixtures inside public/build, which is
// gitignored whole, so they never reach the checked-in tree.
const fixtureBuildDirectory = "build/feature-test"

// viteTestManifest mirrors what `vite build` writes for this app: the Inertia
// entry with its own CSS, a static CSS entry, and a vendor chunk the entry
// imports which itself imports a shared chunk.
const viteTestManifest = `{
  "resources/inertia/app.ts": {"file": "assets/app-4ed993c7.js", "src": "resources/inertia/app.ts", "isEntry": true, "css": ["assets/app-9a1b2c3d.css"], "imports": ["_vendor-1f2e3d4c.js"]},
  "resources/css/app.css": {"file": "assets/site-0a0b0c0d.css", "src": "resources/css/app.css", "isEntry": true},
  "_vendor-1f2e3d4c.js": {"file": "assets/vendor-1f2e3d4c.js", "css": ["assets/vendor-aabbccdd.css"], "imports": ["_shared-9c8b7a6f.js"]},
  "_shared-9c8b7a6f.js": {"file": "assets/shared-9c8b7a6f.js"}
}`

// ViteTestSuite covers the vite module from boot to render: the provider wiring,
// the template functions, vite:inputs, and the failures an app author hits.
type ViteTestSuite struct {
	suite.Suite
	tests.TestCase
	// What the singleton held before any case pointed it at fixtures.
	originalBuildDirectory string
	originalHotFile        string
}

func TestViteTestSuite(t *testing.T) {
	suite.Run(t, &ViteTestSuite{})
}

// SetupSuite snapshots the boot-time config before any case can change or
// restore it, so a broken restore fails here instead of hiding until later.
func (s *ViteTestSuite) SetupSuite() {
	s.originalBuildDirectory = facades.Vite().BuildDirectory()
	s.originalHotFile = facades.Vite().HotFile()

	s.Equal("build", s.originalBuildDirectory)
	s.Equal(path.Base("public", "hot"), s.originalHotFile)
}

func (s *ViteTestSuite) SetupTest() {
	s.Require().NoError(os.MkdirAll(s.fixturePath(), 0o755))
}

func (s *ViteTestSuite) TearDownTest() {
	// The shared singleton is pointed at fixtures during a case, so put back the
	// values it had before the suite started.
	vite, ok := facades.Vite().(*goravelvite.Vite)
	s.Require().True(ok)
	vite.UseBuildDirectory(s.originalBuildDirectory).UseHotFile(s.originalHotFile)

	s.Require().NoError(os.RemoveAll(s.fixturePath()))
}

func (s *ViteTestSuite) fixturePath(parts ...string) string {
	return path.Base(append([]string{"public", fixtureBuildDirectory}, parts...)...)
}

// useFixtureVite points the shared Vite at the fixture build directory and at a
// hot file in a per-test temporary directory. The Use* setters are not on the
// contract on purpose (the instance is configured from config/vite.go), so the
// concrete type is asserted first.
func (s *ViteTestSuite) useFixtureVite() *goravelvite.Vite {
	vite, ok := facades.Vite().(*goravelvite.Vite)
	s.Require().True(ok, "the container should bind the framework *vite.Vite")

	hotFile := filepath.Join(s.T().TempDir(), "hot")
	vite.UseBuildDirectory(fixtureBuildDirectory).UseHotFile(hotFile)

	return vite
}

func (s *ViteTestSuite) writeFile(name, content string) {
	s.Require().NoError(os.MkdirAll(filepath.Dir(name), 0o755))
	s.Require().NoError(os.WriteFile(name, []byte(content), 0o644))
}

func (s *ViteTestSuite) writeManifest(manifest string) {
	s.writeFile(s.fixturePath(".vite", "manifest.json"), manifest)
}

// TestFuncs_Registered pins the provider wiring: the three template functions
// are in the view set the provider booted with.
func (s *ViteTestSuite) TestFuncs_Registered() {
	funcs := facades.View().Funcs()
	for _, name := range []string{"vite", "viteAsset", "viteReactRefresh"} {
		_, ok := funcs[name]
		s.True(ok, "the vite provider should register %q", name)
	}
}

func (s *ViteTestSuite) TestRender() {
	tests := []struct {
		name  string
		setup func()
		check func(html string)
	}{
		{
			name: "should load from the dev server while the hot file exists",
			setup: func() {
				vite := s.useFixtureVite()
				s.writeFile(vite.HotFile(), "http://localhost:5173\n")
			},
			check: func(html string) {
				s.Contains(html, `<script type="module" src="http://localhost:5173/@vite/client"></script>`)
				s.Contains(html, `<script type="module" src="http://localhost:5173/resources/inertia/app.ts"></script>`)
				s.Contains(html, `href="http://localhost:5173/resources/inertia/app.ts"`)
				s.Equal(1, strings.Count(html, "@vite/client"))
				s.NotContains(html, "/build/")
			},
		},
		{
			name: "should load the built files from the manifest",
			setup: func() {
				s.useFixtureVite()
				s.writeManifest(viteTestManifest)
			},
			check: func(html string) {
				s.NotContains(html, "@vite/client")
				s.Contains(html, `<link rel="modulepreload" href="/build/feature-test/assets/vendor-1f2e3d4c.js">`)
				s.Contains(html, `<link rel="stylesheet" href="/build/feature-test/assets/app-9a1b2c3d.css">`)
				s.Contains(html, `<script type="module" src="/build/feature-test/assets/app-4ed993c7.js"></script>`)
				s.Contains(html, `href="/build/feature-test/assets/app-4ed993c7.js"`)
				s.Less(strings.Index(html, `rel="stylesheet"`), strings.Index(html, `<script type="module" src=`))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			html, err := facades.View().Make("vite.tmpl").Render()
			s.Require().NoError(err)

			tt.check(html)
		})
	}
}

// TestTags_SharedChunksAndOrder renders two entries at once, which the single-entry
// template cannot: the shared chunk is preloaded once, not once per entry.
func (s *ViteTestSuite) TestTags_SharedChunksAndOrder() {
	vite := s.useFixtureVite()
	s.writeManifest(viteTestManifest)

	tags, err := vite.Tags("resources/css/app.css", "resources/inertia/app.ts")
	s.Require().NoError(err)
	s.Equal(template.HTML(
		`<link rel="modulepreload" href="/build/feature-test/assets/vendor-1f2e3d4c.js">`+
			`<link rel="modulepreload" href="/build/feature-test/assets/shared-9c8b7a6f.js">`+
			`<link rel="stylesheet" href="/build/feature-test/assets/site-0a0b0c0d.css">`+
			`<link rel="stylesheet" href="/build/feature-test/assets/app-9a1b2c3d.css">`+
			`<link rel="stylesheet" href="/build/feature-test/assets/vendor-aabbccdd.css">`+
			`<script type="module" src="/build/feature-test/assets/app-4ed993c7.js"></script>`,
	), tags)
}

func (s *ViteTestSuite) TestInputsCommand() {
	output, err := s.CaptureArtisanOutput("vite:inputs")
	s.Require().NoError(err)

	// Field by field, so a vite call added to another view does not read as a
	// broken command.
	s.Contains(output, `"inputs":["resources/inertia/app.ts"]`)
	s.Contains(output, `"assets":["resources/inertia/app.ts"]`)
	s.Contains(output, `"build_directory":"build"`)
}

// TestErrors covers the three failures in the order a build goes wrong: no
// manifest yet, a manifest that does not parse, and an entry the manifest lacks.
// Each case builds on the file the previous one wrote.
func (s *ViteTestSuite) TestErrors() {
	tests := []struct {
		name        string
		manifest    string
		entry       string
		expectError error
	}{
		{
			name:        "should report a missing manifest",
			entry:       "resources/inertia/app.ts",
			expectError: frameworkerrors.ViteManifestNotFound,
		},
		{
			name:        "should report an unparseable manifest",
			manifest:    "{not json",
			entry:       "resources/inertia/app.ts",
			expectError: frameworkerrors.ViteManifestInvalid,
		},
		{
			name:        "should report an entry missing from the manifest",
			manifest:    viteTestManifest,
			entry:       "resources/js/missing.js",
			expectError: frameworkerrors.ViteEntryNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			vite := s.useFixtureVite()
			if tt.manifest != "" {
				s.writeManifest(tt.manifest)
			}

			_, err := vite.Tags(tt.entry)
			s.Require().ErrorIs(err, tt.expectError)
		})
	}
}
