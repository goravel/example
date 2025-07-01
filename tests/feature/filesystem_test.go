package feature

import (
	"context"
	"fmt"
	"goravel/tests"
	"os"
	"testing"

	"github.com/goravel/framework/contracts/filesystem"
	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/facades"
	supportdocker "github.com/goravel/framework/support/docker"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/suite"
)

type FilesystemTestSuite struct {
	suite.Suite
	tests.TestCase
	minioDocker *testingdocker.ImageDriver
	drivers     []string
}

func TestFilesystemTestSuite(t *testing.T) {
	suite.Run(t, &FilesystemTestSuite{})
}

func (s *FilesystemTestSuite) SetupSuite() {
	s.drivers = []string{
		"",
		"local",
		"public",
		"s3",
		"cos",
		"oss",
		"cloudinary",
		"minio",
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY_ID")
	minioSecretKey := os.Getenv("MINIO_ACCESS_KEY_SECRET")
	minioBucket := os.Getenv("MINIO_BUCKET")

	docker := testingdocker.NewImageDriver(contractsdocker.Image{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ACCESS_KEY=" + minioAccessKey,
			"MINIO_SECRET_KEY=" + minioSecretKey,
		},
		ExposedPorts: []string{
			"9000",
		},
	})
	err := docker.Build()
	if err != nil {
		panic(err)
	}

	config := docker.Config()
	endpoint := fmt.Sprintf("127.0.0.1:%s", supportdocker.ExposedPort(config.ExposedPorts, "9000"))
	facades.Config().Add("filesystems.disks.minio.endpoint", endpoint)
	facades.Config().Add("filesystems.disks.minio.url", fmt.Sprintf("http://%s/%s", endpoint, minioBucket))
	// facades.App().Refresh()

	if err := docker.Ready(func() error {
		client, err := minio.New(endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		})
		if err != nil {
			return err
		}
		if err := client.MakeBucket(context.Background(), minioBucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}

		policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Action": [
                    "s3:GetObject",
                    "s3:PutObject"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + minioBucket + `/*"
                ]
            },
            {
                "Action": [
                    "s3:ListBucket"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + minioBucket + `"
                ]
            }
        ]
    }`

		if err := client.SetBucketPolicy(context.Background(), minioBucket, policy); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	s.minioDocker = docker
}

func (s *FilesystemTestSuite) SetupTest() {
}

func (s *FilesystemTestSuite) TearDownSuite() {
	s.NoError(s.minioDocker.Shutdown())
}

func (s *FilesystemTestSuite) TestPutAndGet() {
	for _, driver := range s.drivers {
		var storage filesystem.Driver
		if driver == "" {
			storage = facades.Storage()
		} else {
			storage = facades.Storage().Disk(driver)
		}

		s.NoError(storage.Put("test.txt", "test"))
		content, err := storage.Get("test.txt")

		s.NoError(err)
		s.Equal("test", content)

		s.NoError(storage.Delete("test.txt"))
	}
}
