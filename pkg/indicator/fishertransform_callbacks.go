// Code generated by "callbackgen -type FisherTransform"; DO NOT EDIT.

package indicator

import ()

func (inc *FisherTransform) OnUpdate(cb func(value float64)) {
	inc.UpdateCallbacks = append(inc.UpdateCallbacks, cb)
}

func (inc *FisherTransform) EmitUpdate(value float64) {
	for _, cb := range inc.UpdateCallbacks {
		cb(value)
	}
}
