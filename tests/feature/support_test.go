package feature

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
)

func TestPathResource(t *testing.T) {
	resourcePath := path.Resource()
	fmt.Println(resourcePath)
	assert.True(t, strings.HasPrefix(resourcePath, "/"))
	assert.True(t, strings.HasSuffix(resourcePath, "/resources"))

	resourcePath = path.Resource("test.txt")
	fmt.Println(resourcePath)
	assert.True(t, strings.HasPrefix(resourcePath, "/"))
	assert.True(t, strings.HasSuffix(resourcePath, "/resources/test.txt"))
}
