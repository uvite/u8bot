// Code generated by "callbackgen -type A34"; DO NOT EDIT.

package fmaker

import ()

func (inc *A34) OnUpdate(cb func(val float64)) {
	inc.UpdateCallbacks = append(inc.UpdateCallbacks, cb)
}

func (inc *A34) EmitUpdate(val float64) {
	for _, cb := range inc.UpdateCallbacks {
		cb(val)
	}
}
