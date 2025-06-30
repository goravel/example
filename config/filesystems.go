package config

import (
	cloudinaryfacades "github.com/goravel/cloudinary/facades"
	cosfacades "github.com/goravel/cos/facades"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/path"
	miniofacades "github.com/goravel/minio/facades"
)

func init() {
	config := facades.Config()
	config.Add("filesystems", map[string]any{
		// Default Filesystem Disk
		//
		// Here you may specify the default filesystem disk that should be used
		// by the framework. The "local" disk, as well as a variety of cloud
		// based disks are available to your application. Just store away!
		"default": config.Env("FILESYSTEM_DISK", "local"),

		// Filesystem Disks
		//
		// Here you may configure as many filesystem "disks" as you wish, and you
		// may even configure multiple disks of the same driver. Defaults have
		// been set up for each driver as an example of the required values.
		//
		// Supported Drivers: "local", "custom"
		"disks": map[string]any{
			"local": map[string]any{
				"driver": "local",
				"root":   path.Storage("app"),
			},
			"public": map[string]any{
				"driver": "local",
				"root":   path.Storage("app/public"),
				"url":    config.Env("APP_URL", "").(string) + "/storage",
			},
			"cos": map[string]any{
				"driver": "custom",
				"key":    config.Env("TENCENT_ACCESS_KEY_ID"),
				"secret": config.Env("TENCENT_ACCESS_KEY_SECRET"),
				"url":    config.Env("TENCENT_URL"),
				"via": func() (filesystem.Driver, error) {
					return cosfacades.Cos("cos") // The `cos` value is the `disks` key
				},
			},
			"cloudinary": map[string]any{
				"driver": "custom",
				"cloud":  config.Env("CLOUDINARY_CLOUD"),
				"key":    config.Env("CLOUDINARY_ACCESS_KEY_ID"),
				"secret": config.Env("CLOUDINARY_ACCESS_KEY_SECRET"),
				"via": func() (filesystem.Driver, error) {
					return cloudinaryfacades.Cloudinary("cloudinary") // The `cloudinary` value is the `disks` key
				},
			},
			"minio": map[string]any{
				"driver":   "custom",
				"key":      config.Env("MINIO_ACCESS_KEY_ID"),
				"secret":   config.Env("MINIO_ACCESS_KEY_SECRET"),
				"region":   config.Env("MINIO_REGION"),
				"bucket":   config.Env("MINIO_BUCKET"),
				"url":      config.Env("MINIO_URL"),
				"endpoint": config.Env("MINIO_ENDPOINT"),
				"ssl":      config.Env("MINIO_SSL", false),
				"via": func() (filesystem.Driver, error) {
					return miniofacades.Minio("minio") // The `minio` value is the `disks` key
				},
			},
		},
	})
}
