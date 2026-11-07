package feature

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// devServerPort is not 5173 so the suite runs next to a developer's own
// `npm run dev`; vite.config.ts keeps 5173 for the real thing.
const devServerPort = "5180"

// vitePluginInstalled is what npm writes for the published package.
var vitePluginInstalled = path.Base("node_modules", "goravel-vite-plugin", "package.json")

// viteManifestChunk is the part of a manifest.json entry the framework reads.
type viteManifestChunk struct {
	File    string `json:"file"`
	IsEntry bool   `json:"isEntry"`
}

// VitePluginTestSuite drives the published goravel-vite-plugin the way an
// application author does — `npm run build` and `npm run dev` — and then
// resolves the same entry and asset through the framework. config/vite.go is the
// only place the paths are set.
type VitePluginTestSuite struct {
	suite.Suite
	tests.TestCase
	root string
}

func TestVitePluginTestSuite(t *testing.T) {
	suite.Run(t, &VitePluginTestSuite{})
}

// SetupSuite installs the frontend packages once when they are missing, so the
// suite runs on a fresh checkout instead of skipping. CI installs them in its
// own step, so this only pays off locally.
func (s *VitePluginTestSuite) SetupSuite() {
	// Absolute (path.Base resolves "../.." against the test package dir), so a
	// process started with it does not depend on the working directory.
	s.root = path.Base()

	if file.Exists(vitePluginInstalled) {
		return
	}
	if _, err := exec.LookPath("npm"); err != nil {
		s.T().Skip("npm is not installed, so the goravel-vite-plugin cases cannot run")
	}

	result := facades.Process().Path(s.root).Timeout(10*time.Minute).Run("npm", "ci")
	s.Require().False(result.Failed(), "npm ci in %s failed:\n%s\n%s", s.root, result.Output(), result.ErrorOutput())
}

func (s *VitePluginTestSuite) TearDownTest() {
	s.Require().NoError(file.Remove(path.Public(facades.Vite().BuildDirectory())))
	s.Require().NoError(file.Remove(facades.Vite().HotFile()))
}

func (s *VitePluginTestSuite) readManifest() map[string]viteManifestChunk {
	manifestPath := path.Public(facades.Vite().BuildDirectory(), facades.Vite().ManifestFilename())
	content, err := os.ReadFile(manifestPath)
	s.Require().NoError(err, manifestPath)

	var manifest map[string]viteManifestChunk
	s.Require().NoError(json.Unmarshal(content, &manifest))

	return manifest
}

// TestBuild_ProducesTheManifestTheFrameworkReads runs the build an author runs
// and then resolves the same entry and asset through the framework: the entries
// came from the views, the files are where config/vite.go points, and the file a
// view only names through viteAsset is built.
func (s *VitePluginTestSuite) TestBuild_ProducesTheManifestTheFrameworkReads() {
	result := facades.Process().Path(s.root).Timeout(5*time.Minute).Run("npm", "run", "build")
	s.Require().False(result.Failed(), result.Output()+result.ErrorOutput())

	// The entries came from the application, not from a list in vite.config.ts.
	s.True(result.SeeInOutput("artisan vite:inputs"), result.Output())

	// vite.config.ts leaves the paths and the entries to the plugin, so none of
	// its "you are overriding what the application reads" warnings appear
	// (Vite's logger sends warnings to stderr, hence both streams).
	output := result.Output() + result.ErrorOutput()
	s.NotContains(output, "is set in your Vite config")
	s.NotContains(output, "the application reads it from")
	s.NotContains(output, "is not the application's build directory")

	manifest := s.readManifest()
	buildDirectory := facades.Vite().BuildDirectory()

	entry, ok := manifest["resources/inertia/app.ts"]
	s.Require().True(ok, "the plugin should build the entry the views reference")
	s.True(entry.IsEntry)

	// The file nothing imports, named only by {{ viteAsset }} in a view.
	logo, ok := manifest["resources/images/logo.svg"]
	s.Require().True(ok, "the plugin should build the file a view only names through viteAsset")
	s.NotEmpty(logo.File)

	for name, chunk := range manifest {
		s.True(file.Exists(path.Public(buildDirectory, chunk.File)), "%s: %s", name, chunk.File)
	}

	tags, err := facades.Vite().Tags("resources/inertia/app.ts")
	s.Require().NoError(err)
	s.Contains(string(tags), `<script type="module" src="/`+buildDirectory+`/`+entry.File+`"></script>`)

	url, err := facades.Vite().Asset("resources/images/logo.svg")
	s.Require().NoError(err)
	s.Equal("/"+buildDirectory+"/"+logo.File, url)
	s.True(file.Exists(path.Public(buildDirectory, logo.File)), logo.File)
}

// TestDevServer_AnnouncesTheHotFile is the other half: while the dev server runs
// the plugin writes its URL to vite.hot_file, the framework points every tag and
// every asset there, and stopping the server removes the file again.
func (s *VitePluginTestSuite) TestDevServer_AnnouncesTheHotFile() {
	if runtime.GOOS == "windows" {
		// Process.Stop cannot deliver a graceful signal to an npm child there, so
		// the server is killed and leaves its hot file (and possibly a listener)
		// behind.
		s.T().Skip("vite's shutdown handler does not run when the process is killed on Windows")
	}

	hotFile := facades.Vite().HotFile()
	s.Require().NoError(file.Remove(hotFile))

	// The npm script an author runs; npm forwards SIGTERM to vite, whose shutdown
	// handler removes the hot file. The plugin refuses to serve when CI is set,
	// which every CI run is.
	running, err := facades.Process().
		Path(s.root).
		Env(map[string]string{"GORAVEL_BYPASS_ENV_CHECK": "1"}).
		Start("npm", "run", "dev", "--", "--port", devServerPort)
	s.Require().NoError(err)
	defer func() {
		_ = running.Stop(10 * time.Second)
		running.Wait()
	}()

	// Discovery runs `go run .` first, which compiles the application.
	var url string
	s.Require().Eventually(func() bool {
		content, err := os.ReadFile(hotFile)
		if err != nil {
			return false
		}
		url = strings.TrimSpace(string(content))

		return url != ""
	}, 120*time.Second, 100*time.Millisecond, "the plugin should write %s:\n%s", hotFile, running.Output())

	s.Regexp(`^http://(localhost|127\.0\.0\.1):`+devServerPort+`$`, url)
	s.True(facades.Vite().IsRunningHot())

	tags, err := facades.Vite().Tags("resources/inertia/app.ts")
	s.Require().NoError(err)
	got := string(tags)
	s.Contains(got, `<script type="module" src="`+url+`/@vite/client"></script>`)
	s.Contains(got, `<script type="module" src="`+url+`/resources/inertia/app.ts"></script>`)

	// The asset URL comes from the dev server too — no manifest is read while it runs.
	asset, err := facades.Vite().Asset("resources/images/logo.svg")
	s.Require().NoError(err)
	s.Equal(url+"/resources/images/logo.svg", asset)

	s.Require().NoError(running.Stop(10 * time.Second))
	running.Wait()
	s.Require().Eventually(func() bool {
		return !file.Exists(hotFile)
	}, 20*time.Second, 100*time.Millisecond,
		"the plugin should remove %s when the server stops:\n%s", hotFile, running.Output())
}
