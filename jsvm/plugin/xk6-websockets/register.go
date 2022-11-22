// Package websockets exist just to register the websockets extension
package websockets

import (
	"github.com/uvite/u8/js/modules"
	"github.com/uvite/u8/plugin/xk6-websockets/websockets"
)

func init() {
	modules.Register("k6/x/websockets", new(websockets.RootModule))
}
