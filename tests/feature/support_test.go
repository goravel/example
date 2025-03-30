package feature

import (
	"strings"
	"testing"

	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
)

func TestPathResource(t *testing.T) {
	resourcePath := path.Resource()
	assert.True(t, strings.HasPrefix(resourcePath, "/"))
	assert.True(t, strings.HasSuffix(resourcePath, "tests/feature/resources"))

	resourcePath = path.Resource("test.txt")
	assert.True(t, strings.HasPrefix(resourcePath, "/"))
	assert.True(t, strings.HasSuffix(resourcePath, "tests/feature/resources/test.txt"))
}
