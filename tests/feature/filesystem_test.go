package feature

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/filesystem"
	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	supportdocker "github.com/goravel/framework/support/docker"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

type FilesystemTestSuite struct {
	suite.Suite
	tests.TestCase
	minioDocker contractsdocker.ImageDriver
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
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		s.drivers = append(s.drivers, "s3")
	}
	if os.Getenv("ALIYUN_ACCESS_KEY_ID") != "" {
		s.drivers = append(s.drivers, "oss")
	}
	if os.Getenv("TENCENT_ACCESS_KEY_ID") != "" {
		s.drivers = append(s.drivers, "cos")
	}
	if os.Getenv("MINIO_ACCESS_KEY_ID") != "" {
		s.drivers = append(s.drivers, "minio")
		s.minioDocker = initMinio()
	}

	fmt.Printf("testing filesystem drivers: %v\n", s.drivers)
}

func (s *FilesystemTestSuite) SetupTest() {
}

// TearDownTest heals a fake left enabled by an earlier test: Fake on the
// default disk takes over the driver the real disk replaced, and its Reset
// restores it, so the next test sees the configured disk again.
func (s *FilesystemTestSuite) TearDownTest() {
	fake := facades.Storage().Fake()
	s.NoError(fake.Reset())
}

func (s *FilesystemTestSuite) TearDownSuite() {
	if s.minioDocker != nil {
		s.NoError(s.minioDocker.Shutdown())
	}
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

func (s *FilesystemTestSuite) TestFakeDoesNotWriteIntoTheRealDisk() {
	disk := facades.Config().GetString("filesystems.default")
	realRoot := facades.Config().GetString("filesystems.disks." + disk + ".root")

	fake := facades.Storage().Fake()
	defer func() { s.NoError(fake.Reset()) }()

	s.NoError(facades.Storage().Put("avatars/1.png", "image"))
	s.NoError(facades.Storage().Disk(disk).Put("avatars/2.png", "image"))

	// The facade and the named disk both resolve to the fake...
	s.NotEqual(filepath.Clean(realRoot), filepath.Clean(fake.Path("")))
	s.True(fake.AssertExists("avatars/1.png", "avatars/2.png", "avatars"))
	content, err := fake.Get("avatars/1.png")
	s.NoError(err)
	s.Equal("image", content)

	// ...and the real storage directory was never written to.
	s.NoFileExists(filepath.Join(realRoot, "avatars", "1.png"))
	s.NoFileExists(filepath.Join(realRoot, "avatars", "2.png"))

	s.False(fake.AssertExists("avatars/1.png", "avatars/3.png"))
	s.True(fake.AssertMissing("avatars/3.png", "other"))
	s.False(fake.AssertMissing("avatars/3.png", "avatars/1.png"))
}

func (s *FilesystemTestSuite) TestFakeCountsAndDirectories() {
	fake := facades.Storage().Fake()
	defer func() { s.NoError(fake.Reset()) }()

	s.True(fake.AssertCount("", 0))

	s.NoError(fake.Put("a.txt", "a"))
	s.NoError(fake.Put("nested/b.txt", "b"))
	s.NoError(fake.MakeDirectory("empty"))

	s.True(fake.AssertCount("", 1))
	s.True(fake.AssertCount("", 2, true))
	s.True(fake.AssertCount("nested", 1))
	s.False(fake.AssertCount("missing", 0))
	s.False(fake.AssertDirectoryEmpty(""))
	s.True(fake.AssertDirectoryEmpty("empty"))
}

func (s *FilesystemTestSuite) TestFakeKeepsTheDiskUrl() {
	fake := facades.Storage().Fake("public")
	defer func() { s.NoError(fake.Reset()) }()

	s.Same(fake, facades.Storage().Disk("public"))
	s.Equal(
		facades.Config().GetString("filesystems.disks.public.url")+"/avatars/1.png",
		fake.Url("avatars/1.png"),
	)
}

func (s *FilesystemTestSuite) TestFakeResetRestoresTheRealDisk() {
	disk := facades.Config().GetString("filesystems.default")
	realRoot := facades.Config().GetString("filesystems.disks." + disk + ".root")

	fake := facades.Storage().Fake()
	fakeRoot := fake.Path("")

	s.NoError(fake.Put("tmp/fake.txt", "fake"))
	s.True(fake.AssertExists("tmp/fake.txt"))

	s.NoError(fake.Reset())

	s.NoDirExists(fakeRoot)

	// The facade writes to the configured root again.
	s.NoError(facades.Storage().Put("tmp/real.txt", "real"))
	s.T().Cleanup(func() { s.NoError(facades.Storage().Delete("tmp/real.txt")) })

	s.FileExists(filepath.Join(realRoot, "tmp", "real.txt"))
	s.NoFileExists(filepath.Join(realRoot, "tmp", "fake.txt"))
}

func initMinio() contractsdocker.ImageDriver {
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY_ID")
	minioSecretKey := os.Getenv("MINIO_ACCESS_KEY_SECRET")
	minioBucket := os.Getenv("MINIO_BUCKET")

	docker := testingdocker.NewImageDriver(contractsdocker.Image{
		// MinIO removed its official images from Docker Hub (minio/minio now 404s),
		// so use a maintained community rebuild of the last official release.
		Repository: "coollabsio/minio",
		Tag:        "RELEASE.2025-10-15T17-29-55Z",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ROOT_USER=" + minioAccessKey,
			"MINIO_ROOT_PASSWORD=" + minioSecretKey,
		},
		ExposedPorts: []string{
			"9000",
		},
	}, facades.Process())
	err := docker.Build()
	if err != nil {
		panic(err)
	}

	config := docker.Config()
	endpoint := fmt.Sprintf("127.0.0.1:%s", supportdocker.ExposedPort(config.ExposedPorts, "9000"))
	facades.Config().Add("filesystems.disks.minio.endpoint", endpoint)
	facades.Config().Add("filesystems.disks.minio.url", fmt.Sprintf("http://%s/%s", endpoint, minioBucket))

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

	return docker
}
