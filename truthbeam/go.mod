module github.com/complytime/complybeacon/truthbeam

go 1.24.13

tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen

require (
	github.com/maypok86/otter/v2 v2.3.0
	github.com/ossf/gemara v0.12.1
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.51.0
	go.opentelemetry.io/collector/component/componenttest v0.145.0
	go.opentelemetry.io/collector/config/confighttp v0.145.0
	go.opentelemetry.io/collector/consumer v1.51.0
	go.opentelemetry.io/collector/pdata v1.51.0
	go.opentelemetry.io/collector/processor v1.51.0
	go.opentelemetry.io/collector/processor/processorhelper v0.145.0
	go.opentelemetry.io/collector/processor/processortest v0.145.0
	go.uber.org/zap v1.27.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20251226215517-609e4778396f // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/getkin/kin-openapi v0.132.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.2 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.5.0 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.23 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/speakeasy-api/jsonpath v0.6.0 // indirect
	github.com/speakeasy-api/openapi-overlay v0.10.2 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/client v1.51.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.145.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.51.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.51.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.51.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.51.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.51.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.51.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.51.0 // indirect
	go.opentelemetry.io/collector/confmap v1.51.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.145.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.145.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.145.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.51.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.145.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.51.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.145.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.145.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.145.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.51.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.145.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
