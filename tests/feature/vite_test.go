package feature

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// fixtureHotFile keeps the hot file inside that same directory: it is under
// public, so the framework resolves it, and TearDownTest removes it with the
// rest of the fixture.
const fixtureHotFile = fixtureBuildDirectory + "/hot"

// viteTestManifest mirrors what `vite build` writes for this app: the Inertia
// entry with its own CSS, a static CSS entry, a file only a view names through
// viteAsset, and a vendor chunk the entry imports which itself imports a shared
// chunk.
const viteTestManifest = `{
  "resources/inertia/app.ts": {"file": "assets/app-4ed993c7.js", "src": "resources/inertia/app.ts", "isEntry": true, "css": ["assets/app-9a1b2c3d.css"], "imports": ["_vendor-1f2e3d4c.js"]},
  "resources/css/app.css": {"file": "assets/site-0a0b0c0d.css", "src": "resources/css/app.css", "isEntry": true},
  "resources/images/logo.svg": {"file": "assets/logo-0a0b0c0d.svg", "src": "resources/images/logo.svg"},
  "_vendor-1f2e3d4c.js": {"file": "assets/vendor-1f2e3d4c.js", "css": ["assets/vendor-aabbccdd.css"], "imports": ["_shared-9c8b7a6f.js"]},
  "_shared-9c8b7a6f.js": {"file": "assets/shared-9c8b7a6f.js"}
}`

// ViteTestSuite covers the vite module from boot to render: the provider wiring,
// the template functions, vite:inputs, and the failures an app author hits.
type ViteTestSuite struct {
	suite.Suite
	tests.TestCase
	// The build directory the singleton booted with, and the hot file config
	// declares, both leaves relative to public.
	originalBuildDirectory string
	originalHotFile        string
}

func TestViteTestSuite(t *testing.T) {
	suite.Run(t, &ViteTestSuite{})
}

// SetupSuite snapshots the boot-time config before any case can change or
// restore it, so a broken restore fails here instead of hiding until later.
func (s *ViteTestSuite) SetupSuite() {
	// config/vite.go holds leaves; the instance resolves both inside public.
	s.originalBuildDirectory = facades.Vite().BuildDirectory()
	s.originalHotFile = facades.Config().GetString("vite.hot_file", goravelvite.DefaultHotFile())

	s.Equal("build", s.originalBuildDirectory)
	s.Equal(path.Public(s.originalHotFile), facades.Vite().HotFile())
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
// hot file inside it. Both are leaves relative to public, as config/vite.go
// declares them. The Use* setters are not on the contract on purpose (the
// instance is configured from config/vite.go), so the concrete type is asserted
// first.
func (s *ViteTestSuite) useFixtureVite() *goravelvite.Vite {
	vite, ok := facades.Vite().(*goravelvite.Vite)
	s.Require().True(ok, "the container should bind the framework *vite.Vite")

	vite.UseBuildDirectory(fixtureBuildDirectory).UseHotFile(fixtureHotFile)

	// Every case starts from an empty build, so a manifest or a hot file one case
	// wrote cannot decide what the next one sees.
	s.Require().NoError(os.RemoveAll(s.fixturePath(".vite")))
	s.Require().NoError(os.RemoveAll(vite.HotFile()))

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
				s.Contains(html, `href="http://localhost:5173/resources/images/logo.svg"`)
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
				s.Contains(html, `<link rel="icon" href="/build/feature-test/assets/logo-0a0b0c0d.svg" />`)
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
	// broken command. The config seeds the same two paths, so the dedup is what
	// keeps each list at one entry.
	s.Contains(output, `"inputs":["resources/inertia/app.ts"]`)
	s.Contains(output, `"assets":["resources/images/logo.svg"]`)
	s.Contains(output, `"build_directory":"build"`)
	s.NotContains(output, ":null")
}

// TestInputsCommand_SeededEntries pins the entries config declares: vite.inputs
// and vite.assets are merged with the view calls, then deduplicated and sorted,
// so an entry named in both places is built once.
func (s *ViteTestSuite) TestInputsCommand_SeededEntries() {
	scope, err := tests.OverrideConfig(map[string]any{
		"vite.inputs": []string{"resources/js/admin.ts", "resources/inertia/app.ts"},
		"vite.assets": []string{"resources/images/logo.svg", "resources/images/favicon.svg"},
	})
	s.Require().NoError(err)
	defer func() { s.Require().NoError(scope.Restore()) }()

	output, err := s.CaptureArtisanOutput("vite:inputs")
	s.Require().NoError(err)

	// Merged with the view calls (app.ts, logo.svg), once each, in sorted order.
	s.Contains(output, `"inputs":["resources/inertia/app.ts","resources/js/admin.ts"]`)
	s.Contains(output, `"assets":["resources/images/favicon.svg","resources/images/logo.svg"]`)

	// The command sorts its own copy; the value config holds keeps its order.
	s.Equal(
		[]string{"resources/js/admin.ts", "resources/inertia/app.ts"},
		facades.Config().GetStringSlice("vite.inputs", []string{}),
	)
}

// TestInputs_Paths pins the handshake goravel-vite-plugin reads: it builds into
// the paths reported here. The build directory is the leaf config/vite.go sets;
// the hot file is reported relative to the project root, so "hot" inside public
// still reads "public/hot".
func (s *ViteTestSuite) TestInputs_Paths() {
	output, err := s.CaptureArtisanOutput("vite:inputs")
	s.Require().NoError(err)

	// Artisan colours the line, so decode from the object rather than the whole
	// output; the decoder stops at the end of it.
	start := strings.Index(output, "{")
	s.Require().NotEqual(-1, start, output)

	var inputs struct {
		Paths struct {
			PublicDirectory string `json:"public_directory"`
			BuildDirectory  string `json:"build_directory"`
			HotFile         string `json:"hot_file"`
			Manifest        string `json:"manifest"`
		} `json:"paths"`
	}
	s.Require().NoError(json.NewDecoder(strings.NewReader(output[start:])).Decode(&inputs))

	s.Equal("public", inputs.Paths.PublicDirectory)
	s.Equal("build", inputs.Paths.BuildDirectory)
	s.Equal(".vite/manifest.json", inputs.Paths.Manifest)
	// config/vite.go says "hot"; it resolves inside public, and the command
	// reports it relative to the project root so the output is machine-independent.
	s.Equal("hot", facades.Config().GetString("vite.hot_file"))
	s.Equal("public/hot", inputs.Paths.HotFile)
}

// TestManifestHash pins the version an app puts on the files it serves itself:
// the md5 of the build manifest, and "" whenever there is nothing trustworthy to
// hash — no manifest, a broken one, or a dev server that serves unhashed files.
func (s *ViteTestSuite) TestManifestHash() {
	sum := md5.Sum([]byte(viteTestManifest))

	tests := []struct {
		name     string
		setup    func(vite *goravelvite.Vite)
		expected string
	}{
		{
			name: "should hash the manifest a build wrote",
			setup: func(_ *goravelvite.Vite) {
				s.writeManifest(viteTestManifest)
			},
			expected: hex.EncodeToString(sum[:]),
		},
		{
			name:     "should be empty before the first build",
			setup:    func(_ *goravelvite.Vite) {},
			expected: "",
		},
		{
			name: "should be empty while the dev server runs",
			setup: func(vite *goravelvite.Vite) {
				s.writeManifest(viteTestManifest)
				s.writeFile(vite.HotFile(), "http://localhost:5173\n")
			},
			expected: "",
		},
		{
			name: "should be empty for a manifest that does not parse",
			setup: func(_ *goravelvite.Vite) {
				s.writeManifest("{not json")
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			vite := s.useFixtureVite()
			tt.setup(vite)

			s.Equal(tt.expected, vite.ManifestHash())
		})
	}
}

// TestManifestHash_MovesWithTags pins that the version and the tags come from
// one cache: a rebuild that keeps the file size still moves both, because the
// mtime is what invalidates it.
func (s *ViteTestSuite) TestManifestHash_MovesWithTags() {
	vite := s.useFixtureVite()
	s.writeManifest(viteTestManifest)

	first := vite.ManifestHash()
	s.NotEmpty(first)

	_, err := vite.Tags("resources/inertia/app.ts")
	s.Require().NoError(err)
	s.Equal(first, vite.ManifestHash(), "a render must not change the cached hash")

	// Same length, new file name: only the mtime tells the two builds apart.
	rebuilt := strings.Replace(viteTestManifest, "app-4ed993c7.js", "app-44ecf5d8.js", 1)
	s.writeManifest(rebuilt)
	future := time.Now().Add(time.Second)
	s.Require().NoError(os.Chtimes(s.fixturePath(".vite", "manifest.json"), future, future))

	sum := md5.Sum([]byte(rebuilt))
	second := vite.ManifestHash()
	s.Equal(hex.EncodeToString(sum[:]), second)
	s.NotEqual(first, second)

	tags, err := vite.Tags("resources/inertia/app.ts")
	s.Require().NoError(err)
	s.Contains(string(tags), "app-44ecf5d8.js", "the tags must come from the bytes the hash covers")
}

// TestErrors covers the three failures in the order a build goes wrong: no
// manifest yet, a manifest that does not parse, and an entry the manifest lacks.
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
