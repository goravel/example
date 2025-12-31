package main

import (
	"testing"

	"goravel/app/facades"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"

	"goravel/packages/sms"
)

func TestInstallAndUninstallDBDrivers(t *testing.T) {
	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/postgres@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&postgres.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/postgres"))
	assert.False(t, file.Contain(path.Config("database.go"), `"postgres": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/postgres@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&postgres.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/postgres"))
	assert.True(t, file.Contain(path.Config("database.go"), `"postgres": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/mysql@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&mysql.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/mysql"))
	assert.False(t, file.Contain(path.Config("database.go"), `"mysql": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/mysql@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&mysql.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/mysql"))
	assert.True(t, file.Contain(path.Config("database.go"), `"mysql": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/sqlserver@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&sqlserver.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/sqlserver"))
	assert.False(t, file.Contain(path.Config("database.go"), `"sqlserver": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/sqlserver@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&sqlserver.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/sqlserver"))
	assert.True(t, file.Contain(path.Config("database.go"), `"sqlserver": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/sqlite@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&sqlite.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/sqlite"))
	assert.False(t, file.Contain(path.Config("database.go"), `"sqlite": map[string]any{`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/sqlite@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&sqlite.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/sqlite"))
	assert.True(t, file.Contain(path.Config("database.go"), `"sqlite": map[string]any{`))
}

func TestInstallAndUninstallFilesystemDrivers(t *testing.T) {
	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/s3@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&s3.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/s3"))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `"s3": map[string]any{`))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `s3facades "github.com/goravel/s3/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/s3@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&s3.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/s3"))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `"s3": map[string]any{`))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `s3facades "github.com/goravel/s3/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/cos@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&cos.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/cos"))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `"cos": map[string]any{`))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `cosfacades "github.com/goravel/cos/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/cos@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&cos.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/cos"))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `"cos": map[string]any{`))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `cosfacades "github.com/goravel/cos/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/oss@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&oss.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/oss"))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `"oss": map[string]any{`))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `ossfacades "github.com/goravel/oss/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/oss@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&oss.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/oss"))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `"oss": map[string]any{`))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `ossfacades "github.com/goravel/oss/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/minio@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&minio.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/minio"))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `"minio": map[string]any{`))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `miniofacades "github.com/goravel/minio/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/minio@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&minio.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/minio"))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `"minio": map[string]any{`))
	assert.True(t, file.Contain(path.Config("filesystems.go"), `miniofacades "github.com/goravel/minio/facades"`))
}

func TestInstallAndUninstallCacheDrivers(t *testing.T) {
	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/redis@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&redis.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/redis"))
	assert.False(t, file.Contain(path.Config("cache.go"), `"redis": map[string]any{`))
	assert.False(t, file.Contain(path.Config("cache.go"), `github.com/goravel/framework/contracts/cache`))
	assert.False(t, file.Contain(path.Config("cache.go"), `redisfacades "github.com/goravel/redis/facades"`))
	assert.False(t, file.Contain(path.Config("database.go"), `"redis": map[string]any{`))
	assert.False(t, file.Contain(path.Config("queue.go"), `"redis": map[string]any{`))
	assert.False(t, file.Contain(path.Config("session.go"), `"redis": map[string]any{`))
	assert.False(t, file.Contain(path.Config("session.go"), `github.com/goravel/framework/contracts/session`))
	assert.False(t, file.Contain(path.Config("session.go"), `redisfacades "github.com/goravel/redis/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/redis@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&redis.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/redis"))
	assert.True(t, file.Contain(path.Config("cache.go"), `"redis": map[string]any{`))
	assert.True(t, file.Contain(path.Config("cache.go"), `github.com/goravel/framework/contracts/cache`))
	assert.True(t, file.Contain(path.Config("cache.go"), `redisfacades "github.com/goravel/redis/facades"`))
	assert.True(t, file.Contain(path.Config("database.go"), `"redis": map[string]any{`))
	assert.True(t, file.Contain(path.Config("queue.go"), `"redis": map[string]any{`))
	assert.True(t, file.Contain(path.Config("queue.go"), `github.com/goravel/framework/contracts/queue`))
	assert.True(t, file.Contain(path.Config("queue.go"), `redisfacades "github.com/goravel/redis/facades"`))
	assert.True(t, file.Contain(path.Config("session.go"), `"redis": map[string]any{`))
	assert.True(t, file.Contain(path.Config("session.go"), `github.com/goravel/framework/contracts/session`))
	assert.True(t, file.Contain(path.Config("session.go"), `redisfacades "github.com/goravel/redis/facades"`))
}

func TestInstallAndUninstallHttpDrivers(t *testing.T) {
	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/gin@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&gin.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/gin"))
	assert.False(t, file.Contain(path.Config("http.go"), `"gin": map[string]any{`))
	assert.False(t, file.Contain(path.Config("http.go"), `ginfacades "github.com/goravel/gin/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/gin@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&gin.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/gin"))
	assert.True(t, file.Contain(path.Config("http.go"), `"gin": map[string]any{`))
	assert.True(t, file.Contain(path.Config("http.go"), `github.com/goravel/framework/contracts/route`))
	assert.True(t, file.Contain(path.Config("http.go"), `ginfacades "github.com/goravel/gin/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/fiber@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&fiber.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/fiber"))
	assert.False(t, file.Contain(path.Config("http.go"), `"fiber": map[string]any{`))
	assert.False(t, file.Contain(path.Config("http.go"), `fiberfacades "github.com/goravel/fiber/facades"`))

	assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/fiber@master"))
	assert.True(t, file.Contain(path.Config("app.go"), "&fiber.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/fiber"))
	assert.True(t, file.Contain(path.Config("http.go"), `"fiber": map[string]any{`))
	assert.True(t, file.Contain(path.Config("http.go"), `github.com/goravel/framework/contracts/route`))
	assert.True(t, file.Contain(path.Config("http.go"), `fiberfacades "github.com/goravel/fiber/facades"`))
}

func TestInstallAndUninstallLocalPackage(t *testing.T) {
	assert.NoError(t, facades.Artisan().Call("make:package example"))
	assert.True(t, file.Exists(path.Base("packages", "example")))
	assert.True(t, file.Exists(path.Base("packages", "example", "setup", "setup.go")))

	assert.NoError(t, facades.Artisan().Call("package:install goravel/packages/example"))
	assert.True(t, file.Contain(path.Config("app.go"), "&example.ServiceProvider{},"))
	assert.True(t, file.Contain(path.Config("app.go"), "goravel/packages/example"))

	assert.NoError(t, facades.Artisan().Call("package:uninstall goravel/packages/example"))
	assert.False(t, file.Contain(path.Config("app.go"), "&example.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "goravel/packages/example"))

	assert.NoError(t, file.Remove(path.Base("packages", "example")))
}

func TestInstallAndPublishAndUninstallLocalPackage(t *testing.T) {
	assert.NoError(t, facades.Artisan().Call("package:install ./packages/sms --no-ansi"))
	assert.True(t, file.Contain(path.Config("app.go"), "goravel/packages/sms"))
	assert.True(t, file.Contain(path.Config("app.go"), "&sms.ServiceProvider{}"))

	facades.Config().Add("app.providers", append(facades.Config().Get("app.providers").([]foundation.ServiceProvider), &sms.ServiceProvider{}))
	facades.App().Refresh()

	assert.NoError(t, facades.Artisan().Call("vendor:publish --package=./packages/sms --no-ansi"))
	assert.True(t, file.Exists(path.Config("sms.go")))
	assert.NoError(t, file.Remove(path.Config("sms.go")))
	assert.NoError(t, facades.Artisan().Call("package:uninstall ./packages/sms"))
}
