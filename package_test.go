package main

import (
	"goravel/bootstrap"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
)

func TestInstallAndUninstallDBDrivers(t *testing.T) {
	bootstrap.Boot()

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
	bootstrap.Boot()

	assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/s3@master"))
	assert.False(t, file.Contain(path.Config("app.go"), "&s3.ServiceProvider{},"))
	assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/s3"))
	assert.False(t, file.Contain(path.Config("filesystems.go"), `"s3": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/s3@master"))
	// assert.True(t, file.Contain(path.Config("app.go"), "&s3.ServiceProvider{},"))
	// assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/s3"))
	// assert.True(t, file.Contain(path.Config("filesystems.go"), `"s3": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/cos@master"))
	// assert.False(t, file.Contain(path.Config("app.go"), "&cos.ServiceProvider{},"))
	// assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/cos"))
	// assert.False(t, file.Contain(path.Config("filesystems.go"), `"cos": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/cos@master"))
	// assert.True(t, file.Contain(path.Config("app.go"), "&cos.ServiceProvider{},"))
	// assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/cos"))
	// assert.True(t, file.Contain(path.Config("filesystems.go"), `"cos": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/oss@master"))
	// assert.False(t, file.Contain(path.Config("app.go"), "&oss.ServiceProvider{},"))
	// assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/oss"))
	// assert.False(t, file.Contain(path.Config("filesystems.go"), `"oss": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/postgres@master"))
	// assert.True(t, file.Contain(path.Config("app.go"), "&postgres.ServiceProvider{},"))
	// assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/postgres"))
	// assert.True(t, file.Contain(path.Config("filesystems.go"), `"postgres": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/cloudinary@master"))
	// assert.False(t, file.Contain(path.Config("app.go"), "&cloudinary.ServiceProvider{},"))
	// assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/cloudinary"))
	// assert.False(t, file.Contain(path.Config("filesystems.go"), `"cloudinary": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/cloudinary@master"))
	// assert.True(t, file.Contain(path.Config("app.go"), "&cloudinary.ServiceProvider{},"))
	// assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/cloudinary"))
	// assert.True(t, file.Contain(path.Config("filesystems.go"), `"cloudinary": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:uninstall github.com/goravel/minio@master"))
	// assert.False(t, file.Contain(path.Config("app.go"), "&minio.ServiceProvider{},"))
	// assert.False(t, file.Contain(path.Config("app.go"), "github.com/goravel/minio"))
	// assert.False(t, file.Contain(path.Config("filesystems.go"), `"minio": map[string]any{`))

	// assert.NoError(t, facades.Artisan().Call("package:install github.com/goravel/minio@master"))
	// assert.True(t, file.Contain(path.Config("app.go"), "&minio.ServiceProvider{},"))
	// assert.True(t, file.Contain(path.Config("app.go"), "github.com/goravel/minio"))
	// assert.True(t, file.Contain(path.Config("filesystems.go"), `"minio": map[string]any{`))
}
