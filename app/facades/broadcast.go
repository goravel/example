package facades

import (
	"github.com/goravel/framework/contracts/broadcasting"
)

func Broadcast() broadcasting.Broadcast {
	return App().MakeBroadcast()
}
