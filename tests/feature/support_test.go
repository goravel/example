package feature

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
)

func TestPathResource(t *testing.T) {
	resourcePath := path.Resource()
	assert.True(t, strings.HasPrefix(resourcePath, "/"))
	assert.True(t, strings.HasSuffix(resourcePath, "/resources"))

	resourcePath = path.Resource("test.txt")
	assert.True(t, strings.HasPrefix(resourcePath, "/"))
	assert.True(t, strings.HasSuffix(resourcePath, "/resources/test.txt"))
}

func TestChangeChdir(t *testing.T) {
	info, err := os.Stat("go.mod")
	assert.Nil(t, info)
	assert.Error(t, err)

	fmt.Println(support.RelativePath)
	t.Chdir(support.RelativePath)

	info, err = os.Stat("go.mod")
	assert.NotNil(t, info)
	assert.Nil(t, err)
}
