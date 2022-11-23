// Package timers is here just to register the k6/x/events module
package ta

import (
	"github.com/uvite/u8/js/modules"
	"github.com/uvite/u8/plugin/xk6-ta/ta"
)

func init() {
	modules.Register("k6/x/ta", new(ta.RootModule))
}
