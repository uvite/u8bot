package vm

//
//import (
//	"fmt"
//	"github.com/sirupsen/logrus"
//	"github.com/spf13/afero"
//	"github.com/uvite/u8/js"
//	"github.com/uvite/u8/lib"
//	"github.com/uvite/u8/loader"
//	"github.com/uvite/u8/metrics"
//	"gopkg.in/guregu/null.v3"
//	"net/url"
//	"time"
//)
//
//package main
//
//import (
//"context"
//"fmt"
//"github.com/sirupsen/logrus"
//"github.com/spf13/afero"
//"github.com/uvite/v9/js"
//"github.com/uvite/v9/lib"
//"github.com/uvite/v9/loader"
//"github.com/uvite/v9/metrics"
//"gopkg.in/guregu/null.v3"
//"net/url"
//"time"
//)
//
//func getSimpleRunner(filename, data string, opts ...interface{}) (*js.Runner, error) {
//	var (
//		rtOpts      = lib.RuntimeOptions{CompatibilityMode: null.NewString("base", true)}
//		logger      = logrus.New()
//		fsResolvers = map[string]afero.Fs{"file": afero.NewMemMapFs(), "https": afero.NewMemMapFs()}
//	)
//	for _, o := range opts {
//		switch opt := o.(type) {
//		case afero.Fs:
//			fsResolvers["file"] = opt
//		case map[string]afero.Fs:
//			fsResolvers = opt
//		case lib.RuntimeOptions:
//			rtOpts = opt
//		case *logrus.Logger:
//			logger = opt
//		default:
//			fmt.Println("unknown test option %q", opt)
//		}
//	}
//
//	registry := metrics.NewRegistry()
//	builtinMetrics := metrics.RegisterBuiltinMetrics(registry)
//	return js.New(
//		&lib.InitState{
//			Logger:         logger,
//			RuntimeOptions: rtOpts,
//			BuiltinMetrics: builtinMetrics,
//			Registry:       registry,
//		},
//		&loader.SourceData{
//			URL:  &url.URL{Path: filename, Scheme: "file"},
//			Data: []byte(data),
//		},
//		fsResolvers,
//	)
//}
//func main() {
//
//	r, err := getSimpleRunner("/script.js", `
//			var http = require("k6/http");
//
//			exports.default = function() {
//				var doc =  http.get("https://www.runoob.com/go/go-interfaces.html") ;
//
//				 console.log(doc)
//			}
//		`)
//	//logger := logrus.New()
//	//
//	//bi, err := r.Bundle.Instantiate(logger, 0)
//	//
//	//registry := metrics.NewRegistry()
//	//builtinMetrics := metrics.RegisterBuiltinMetrics(registry)
//	//
//	//root, err := lib.NewGroup("", nil)
//	//
//	//bi.ModuleVUImpl.Status = &lib.State{
//	//	Options: lib.Options{},
//	//	Logger:  logger,
//	//	Group:   root,
//	//	Transport: &http.Transport{
//	//		DialContext: (netext.NewDialer(
//	//			net.Dialer{
//	//				Timeout:   10 * time.Second,
//	//				KeepAlive: 60 * time.Second,
//	//				DualStack: true,
//	//			},
//	//			netext.NewResolver(net.LookupIP, 0, types.DNSfirst, types.DNSpreferIPv4),
//	//		)).DialContext,
//	//	},
//	//	BPool:          bpool.NewBufferPool(1),
//	//	Samples:        make(chan metrics.SampleContainer, 500),
//	//	BuiltinMetrics: builtinMetrics,
//	//	Tags:           lib.NewVUStateTags(registry.RootTagSet()),
//	//}
//	//
//	//ctx, cancel := context.WithCancel(context.Background())
//	//defer cancel()
//	//bi.ModuleVUImpl.Ctx = ctx
//	//v, err := bi.GetCallableExport(consts.DefaultFn)(goja.Undefined())
//	//
//	//fmt.Println(v, err)
//	fmt.Println(err)
//	ch := make(chan metrics.SampleContainer, 1000)
//	initVU, err := r.NewVU(0, 0, ch)
//
//	ctx, cancel := context.WithCancel(context.Background())
//	vu := initVU.Activate(&lib.VUActivationParams{RunContext: ctx})
//	errC := make(chan error)
//	go func() { errC <- vu.RunOnce() }()
//
//	select {
//	case <-time.After(15 * time.Second):
//		cancel()
//		fmt.Println("Test timed out")
//	case err := <-errC:
//		cancel()
//		fmt.Println(err)
//	}
//}
