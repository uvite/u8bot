// +heroku install ./cmd/bbgo

module github.com/c9s/bbgo

go 1.19

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Masterminds/squirrel v1.5.3
	github.com/adshao/go-binance/v2 v2.3.8
	github.com/c-bata/goptuna v0.8.1
	github.com/c9s/requestgen v1.3.0
	github.com/c9s/rockhopper v1.2.2-0.20220617053729-ffdc87df194b
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/codingconcepts/env v0.0.0-20200821220118-a8fbf8d84482
	github.com/davecgh/go-spew v1.1.1
	github.com/evanphx/json-patch/v5 v5.6.0
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.13.0
	github.com/gertd/go-pluralize v0.2.1
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.3
	github.com/go-redis/redis/v8 v8.8.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gofrs/flock v0.8.1
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/jedib0t/go-pretty/v6 v6.3.6
	github.com/jmoiron/sqlx v1.3.4
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v5 v5.0.0-20220201181537-ed2888cfa198
	github.com/leekchan/accounting v0.0.0-20191218023648-17a4ce5f94d4
	github.com/lestrrat-go/file-rotatelogs v2.2.0+incompatible
	github.com/mattn/go-shellwords v1.0.12
	github.com/muesli/clusters v0.0.0-20180605185049-a07a36e67d36
	github.com/muesli/kmeans v0.3.0
	github.com/nats-io/nats.go v1.20.0
	github.com/pkg/errors v0.9.1
	github.com/pocketbase/dbx v1.7.0
	github.com/pocketbase/pocketbase v0.7.10
	github.com/pquerna/otp v1.3.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/robfig/cron/v3 v3.0.0
	github.com/rs/zerolog v1.15.0
	github.com/sajari/regression v1.0.1
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.10.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.4
	github.com/valyala/fastjson v1.5.1
	github.com/wcharczuk/go-chart/v2 v2.1.0
	github.com/webview/webview v0.0.0-20210216142346-e0bfdf0e5d90
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/zserge/lorca v0.1.9
	go.uber.org/multierr v1.8.0
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	golang.org/x/time v0.0.0-20220922220347-f3bd1da661af
	gonum.org/v1/gonum v0.8.2
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/tucnak/telebot.v2 v2.5.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/AlecAivazis/survey/v2 v2.3.6 // indirect
	github.com/VividCortex/ewma v1.1.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-sdk-go v1.44.102 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.16 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.20 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.19 // indirect
	github.com/aws/smithy-go v1.13.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/denisenkom/go-mssqldb v0.12.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/domodwyer/mailyak/v3 v3.3.4 // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.1 // indirect
	github.com/ganigeorgiev/fexpr v0.1.1 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-test/deep v1.0.6 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/gax-go/v2 v2.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/strftime v1.0.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nats-server/v2 v2.9.6 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tebeka/strftime v0.1.3 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel v0.19.0 // indirect
	go.opentelemetry.io/otel/metric v0.19.0 // indirect
	go.opentelemetry.io/otel/trace v0.19.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	gocloud.dev v0.26.0 // indirect
	golang.org/x/exp v0.0.0-20220426173459-3bcf042a4bf5 // indirect
	golang.org/x/image v0.0.0-20220902085622-e7cb96979f69 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220921155015-db77216a4ee9 // indirect
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1 // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.12 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.96.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220920201722-2b89144ce006 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.39.0 // indirect
	modernc.org/ccgo/v3 v3.16.9 // indirect
	modernc.org/libc v1.19.0 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.4.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.19.1 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
)
