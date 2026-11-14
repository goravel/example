package feature

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// vitePluginInstalled is what npm writes for the plugin: a link to the local
// ../vite-plugin checkout until a release lists module entries.
var vitePluginInstalled = path.Base("node_modules", "goravel-vite-plugin", "package.json")

// viteManifestChunk is the part of a manifest.json entry the framework reads.
type viteManifestChunk struct {
	File    string `json:"file"`
	IsEntry bool   `json:"isEntry"`
}

// VitePluginTestSuite drives goravel-vite-plugin the way an
// application author does — `npm run build` and `npm run dev` — and then
// resolves the same entry and asset through the framework. config/vite.go is the
// only place the paths are set.
type VitePluginTestSuite struct {
	suite.Suite
	tests.TestCase
	root string
	// Whether a dev server the developer left running had already announced
	// itself, and what it said, so the suite can put it back.
	hadHotFile       bool
	announcedHotFile []byte
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

	if content, err := os.ReadFile(facades.Vite().HotFile()); err == nil {
		s.hadHotFile = true
		s.announcedHotFile = content
	}

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
	s.Require().NoError(s.releaseHotFile())
}

// releaseHotFile removes the hot file the cases wrote, and puts back the one a
// dev server had announced before the suite took it away.
func (s *VitePluginTestSuite) releaseHotFile() error {
	hotFile := facades.Vite().HotFile()
	if !s.hadHotFile {
		return file.Remove(hotFile)
	}

	return os.WriteFile(hotFile, s.announcedHotFile, 0o644)
}

// freePort asks the OS for a port, so a leftover server cannot make Vite refuse
// to bind and the wait below run out.
func freePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer func() { _ = listener.Close() }()

	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), nil
}

func (s *VitePluginTestSuite) readManifest() map[string]viteManifestChunk {
	manifestPath := path.Public(facades.Vite().BuildDirectory(), facades.Vite().ManifestFilename())
	content, err := os.ReadFile(manifestPath)
	s.Require().NoError(err, manifestPath)

	var manifest map[string]viteManifestChunk
	s.Require().NoError(json.Unmarshal(content, &manifest))

	return manifest
}

// manifestSum is the md5 of the manifest a build wrote: the version an app pins
// on the files it serves itself, so a new build invalidates them together.
func (s *VitePluginTestSuite) manifestSum() string {
	content, err := os.ReadFile(path.Public(facades.Vite().BuildDirectory(), facades.Vite().ManifestFilename()))
	s.Require().NoError(err)

	sum := md5.Sum(content)

	return hex.EncodeToString(sum[:])
}

// TestBuild_ProducesTheManifestTheFrameworkReads runs the build an author runs
// and then resolves the same entry and asset through the framework: the entries
// came from the views, the files are where config/vite.go points, and the file a
// view only names through viteAsset is built.
func (s *VitePluginTestSuite) TestBuild_ProducesTheManifestTheFrameworkReads() {
	// A running dev server answers Tags and Asset with its own URLs; this case
	// reads the manifest, so its hot file goes first.
	s.Require().NoError(file.Remove(facades.Vite().HotFile()))

	result := facades.Process().Path(s.root).Timeout(5*time.Minute).Run("npm", "run", "build")
	s.Require().False(result.Failed(), result.Output()+result.ErrorOutput())

	// The entries came from the application, not from a list in vite.config.ts.
	s.True(result.SeeInOutput("artisan vite:inputs"), result.Output())

	// vite.config.ts leaves the paths and the entries to the plugin, so none of
	// its "you are overriding what the application reads" warnings appear
	// (Vite's logger sends warnings to stderr, hence both streams). The plugin is
	// pinned in package.json because these sentences are its own.
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

	// The modules' ViteEntries() are built although no view names them.
	for _, name := range []string{"modules/blog/resources/js/dashboard.ts", "modules/admin/resources/js/main.ts"} {
		chunk, ok := manifest[name]
		s.Require().True(ok, "the plugin should build the module entry %s", name)
		s.True(chunk.IsEntry)
	}

	for name, chunk := range manifest {
		// The alias table the plugin writes for the framework is not a chunk.
		if name == "goravel:aliases" {
			continue
		}
		s.True(file.Exists(path.Public(buildDirectory, chunk.File)), "%s: %s", name, chunk.File)
	}

	tags, err := facades.Vite().Tags("resources/inertia/app.ts")
	s.Require().NoError(err)
	s.Contains(string(tags), `<script type="module" src="/`+buildDirectory+`/`+entry.File+`"></script>`)

	// The version the app serves matches the manifest this build wrote.
	s.Equal(s.manifestSum(), facades.Vite().ManifestHash())

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

	port, err := freePort()
	s.Require().NoError(err)

	// The npm script an author runs; npm forwards SIGTERM to vite, whose shutdown
	// handler removes the hot file. The plugin refuses to serve when CI is set,
	// which every CI run is.
	running, err := facades.Process().
		Path(s.root).
		Env(map[string]string{"GORAVEL_BYPASS_ENV_CHECK": "1"}).
		Start("npm", "run", "dev", "--", "--port", port)
	s.Require().NoError(err)
	defer func() {
		_ = running.Stop(10 * time.Second)
		running.Wait()
	}()

	// Discovery runs `go run .` first, which compiles the application. A server
	// that exits first (npm error, a port taken in the meantime) fails here
	// instead of waiting out the deadline.
	// The plugin writes JSON: the URL and the alias table, module aliases included.
	var hot struct {
		URL     string            `json:"url"`
		Aliases map[string]string `json:"aliases"`
	}
	deadline := time.After(120 * time.Second)
	poll := time.NewTicker(100 * time.Millisecond)
	defer poll.Stop()
	for hot.URL == "" {
		select {
		case <-running.Done():
			s.FailNow("the dev server exited before writing " + hotFile + ":\n" + running.Wait().Output())
		case <-deadline:
			_ = running.Stop(10 * time.Second)
			s.FailNow("the plugin should have written " + hotFile + ":\n" + running.Wait().Output())
		case <-poll.C:
			// The file may be caught half written, so a read that does not parse is retried.
			if content, err := os.ReadFile(hotFile); err == nil {
				_ = json.Unmarshal(content, &hot)
			}
		}
	}

	url := hot.URL
	s.Regexp(`^http://(localhost|127\.0\.0\.1):`+port+`$`, url)
	s.Equal(map[string]string{
		"@":      "resources/inertia",
		"@blog":  "modules/blog/resources/js",
		"@admin": "modules/admin/resources/js",
	}, hot.Aliases)
	s.True(facades.Vite().IsRunningHot())

	// The dev server serves unhashed files, so a manifest the last build left
	// gives no version while it runs. TearDownTest removes it with the build.
	manifestPath := path.Public(facades.Vite().BuildDirectory(), facades.Vite().ManifestFilename())
	s.Require().NoError(os.MkdirAll(filepath.Dir(manifestPath), 0o755))
	s.Require().NoError(os.WriteFile(manifestPath, []byte("{}"), 0o644))
	s.Empty(facades.Vite().ManifestHash())

	tags, err := facades.Vite().Tags("resources/inertia/app.ts")
	s.Require().NoError(err)
	got := string(tags)
	s.Contains(got, `<script type="module" src="`+url+`/@vite/client"></script>`)
	s.Contains(got, `<script type="module" src="`+url+`/resources/inertia/app.ts"></script>`)

	// A module alias resolves through the table in the hot file.
	tags, err = facades.Vite().Tags("@blog/dashboard.ts")
	s.Require().NoError(err)
	s.Contains(string(tags), `<script type="module" src="`+url+`/modules/blog/resources/js/dashboard.ts"></script>`)

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
