package js

import (
	"github.com/uvite/u8/js/modules"
	"github.com/uvite/u8/js/modules/k6"
	"github.com/uvite/u8/js/modules/k6/crypto"
	"github.com/uvite/u8/js/modules/k6/crypto/x509"
	"github.com/uvite/u8/js/modules/k6/data"
	"github.com/uvite/u8/js/modules/k6/encoding"
	"github.com/uvite/u8/js/modules/k6/execution"
	"github.com/uvite/u8/js/modules/k6/grpc"
	"github.com/uvite/u8/js/modules/k6/html"
	"github.com/uvite/u8/js/modules/k6/http"
	"github.com/uvite/u8/js/modules/k6/metrics"
	"github.com/uvite/u8/js/modules/k6/ws"

	"github.com/grafana/xk6-redis/redis"
	"github.com/grafana/xk6-timers/timers"
	expws "github.com/grafana/xk6-websockets/websockets"
)

func getInternalJSModules() map[string]interface{} {
	return map[string]interface{}{
		"k6": k6.New(),
		//"k6/ta":                      ta.New(),
		"k6/crypto":                  crypto.New(),
		"k6/crypto/x509":             x509.New(),
		"k6/data":                    data.New(),
		"k6/encoding":                encoding.New(),
		"k6/execution":               execution.New(),
		"k6/experimental/redis":      redis.New(),
		"k6/experimental/websockets": &expws.RootModule{},
		"k6/experimental/timers":     timers.New(),
		"k6/net/grpc":                grpc.New(),
		"k6/html":                    html.New(),
		"k6/http":                    http.New(),
		"k6/metrics":                 metrics.New(),
		"k6/ws":                      ws.New(),
	}
}

func getJSModules() map[string]interface{} {
	result := getInternalJSModules()
	external := modules.GetJSModules()

	// external is always prefixed with `k6/x`
	for k, v := range external {
		result[k] = v
	}

	return result
}
