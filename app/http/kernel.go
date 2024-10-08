package http

import (
	"github.com/goravel/framework/contracts/http"
	httpmiddleware "github.com/goravel/framework/http/middleware"
	"github.com/goravel/framework/session/middleware"
)

type Kernel struct {
}

// The application's global HTTP middleware stack.
// These middleware are run during every request to your application.
func (kernel Kernel) Middleware() []http.Middleware {
	return []http.Middleware{
		httpmiddleware.Throttle("global"),
		middleware.StartSession(),
	}
}
