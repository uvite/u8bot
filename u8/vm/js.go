package vm

import (
	"fmt"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"github.com/uvite/u8/js"
	"github.com/uvite/u8/lib"
	"github.com/uvite/u8/metrics"
	"gopkg.in/guregu/null.v3"
	"net/url"
)

func getTestPreInitState(logger *logrus.Logger, rtOpts *lib.RuntimeOptions) *lib.InitState {

	if rtOpts == nil {
		rtOpts = &lib.RuntimeOptions{}
	}
	reg := metrics.NewRegistry()
	return &lib.InitState{
		Logger:         logger,
		RuntimeOptions: *rtOpts,
		Registry:       reg,
		BuiltinMetrics: metrics.RegisterBuiltinMetrics(reg),
	}
}

type JsVm struct {
	Bi *js.BundleInstance
}

func NewJsVm(code []byte) *JsVm {

	//code := `export let options = { vus: 12345 }; export default function() { return options.vus; };`
	arc := &lib.Archive{
		Type:        "js",
		FilenameURL: &url.URL{Scheme: "file", Path: "/script"},
		K6Version:   "1",
		Data:        code,
		Options:     lib.Options{VUs: null.IntFrom(999)},
		PwdURL:      &url.URL{Scheme: "file", Path: "/"},
		Filesystems: nil,
	}

	logger := logrus.New()

	b, err := js.NewBundleFromArchive(getTestPreInitState(logger, nil), arc)
	fmt.Println(err)
	bi, err := b.Instantiate(logger, 0)

	fmt.Println(err)
	//val, err := bi.GetCallableExport(DefaultFn)(goja.Undefined())
	//require.NoError(t, err)
	//fmt.Println(val.Export())
	return &JsVm{
		Bi: bi,
	}

}
func (js *JsVm) GetInit() {
	val, err := js.Bi.GetCallableExport(InitFn)(goja.Undefined())

	fmt.Println(val.Export(), err)
}

func (js *JsVm) Default(f interface{}) {

	//val, err := js.Bi.GetCallableExport(DefaultFn)(goja.Undefined())
	//fn := js.Bi.GetExported(DefaultFn)
	//
	//var fn1 func(s bbgo.SingleExchangeStrategy)
	//if err := js.Bi.Runtime.ExportTo(fn, &fn1); err != nil {
	//
	//}
	//
	//fn1(s)
}

func (js *JsVm) OnBar(kline types.KLine) {
	fn := js.Bi.GetExported(OnBar)

	//fn, err := js.Bi.GetCallableExport(InitFn)(goja.Undefined())
	////v := fn(goja.Undefined(), args)
	//a.GetInit()
	var fn1 func(line types.KLine) string
	if err := js.Bi.Runtime.ExportTo(fn, &fn1); err != nil {

	}
	//
	fn1(kline)
	//fmt.Println(v, err)
}
