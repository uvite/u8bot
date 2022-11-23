package modulestest

import (
	"context"

	"github.com/dop251/goja"
	"github.com/uvite/u8/js/common"
	"github.com/uvite/u8/js/modules"
	"github.com/uvite/u8/lib"
)

var _ modules.VU = &VU{}

// VU is a modules.VU implementation meant to be used within tests
type VU struct {
	CtxField              context.Context
	InitEnvField          *common.InitEnvironment
	StateField            *lib.State
	RuntimeField          *goja.Runtime
	RegisterCallbackField func() func(f func() error)
}

// Context returns internally set field to conform to modules.VU interface
func (m *VU) Context() context.Context {
	return m.CtxField
}

// InitEnv returns internally set field to conform to modules.VU interface
func (m *VU) InitEnv() *common.InitEnvironment {
	m.checkIntegrity()
	return m.InitEnvField
}

// State returns internally set field to conform to modules.VU interface
func (m *VU) State() *lib.State {
	m.checkIntegrity()
	return m.StateField
}

// Runtime returns internally set field to conform to modules.VU interface
func (m *VU) Runtime() *goja.Runtime {
	return m.RuntimeField
}

// RegisterCallback is not really implemented
func (m *VU) RegisterCallback() func(f func() error) {
	return m.RegisterCallbackField()
}

func (m *VU) checkIntegrity() {
	if m.InitEnvField != nil && m.StateField != nil {
		panic("there is a bug in the test: InitEnvField and StateField are not allowed at the same time")
	}
}
