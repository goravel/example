package filters

import (
	"context"
	"fmt"
)

type AppendSuffix struct{}

func (r *AppendSuffix) Signature() string {
	return "append_suffix"
}

func (r *AppendSuffix) Handle(ctx context.Context) any {
	return func(val string, suffix ...string) (string, error) {
		if len(suffix) == 0 {
			return "", fmt.Errorf("append_suffix requires at least one suffix")
		}

		return val + suffix[0], nil
	}
}
