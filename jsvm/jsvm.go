package jsvm

import (
	"context"
	"fmt"
	"github.com/c9s/bbgo/jsvm/genv"
	"github.com/dop251/goja"
	"github.com/spf13/afero"
	"github.com/uvite/u8/js"
	"github.com/uvite/u8/lib"
	"github.com/uvite/u8/metrics"

	_ "github.com/uvite/u8/plugin/xk6-nats"
	_ "github.com/uvite/u8/plugin/xk6-ta"
)

type JsVm struct {
	Runner *js.Runner
	Vu     lib.ActiveVU
	*goja.Runtime
}

func NewJsVm(code []byte) (*JsVm, error) {

	fs := afero.NewOsFs()
	//pwd, _ := os.Getwd()
	//logger := logrus.New()
	//sourceData, err := loader.ReadSource(logger, pwd+"/bbgo.js", pwd, map[string]afero.Fs{"file": fs}, nil)
	//if err != nil {
	//	return nil, fmt.Errorf("couldn't set exported options with merged values: %w", err)
	//
	//}
	rtOpts := lib.RuntimeOptions{Genv: map[string]any{

		"okk": 4444,
	}}

	r, err := genv.GetSimpleRunner("/script.js", fmt.Sprintf(`
import {Nats} from 'k6/x/nats';
import ta from 'k6/x/ta';
import {sleep} from 'k6';

%s
			`, code), fs, rtOpts)

	if err != nil {
		return nil, fmt.Errorf("couldn't set exported options with merged values: %w", err)

	}
	ch := make(chan metrics.SampleContainer, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = r.Setup(ctx, ch)
	cancel()

	initVU, err := r.NewVU(1, 1, ch)

	ctx, cancel = context.WithCancel(context.Background())
	//defer cancel()
	vu := initVU.Activate(&lib.VUActivationParams{RunContext: ctx})
	jsvm := &JsVm{
		Runner: r,
		Vu:     vu,
	}
	jsvm.Runtime = r.Bundle.Vm
	return jsvm, nil

}
