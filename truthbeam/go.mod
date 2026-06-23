module github.com/complytime/complybeacon/truthbeam

go 1.26.4

tool (
	github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	github.com/unbound-force/gaze/cmd/gaze
)

require (
	github.com/gemaraproj/go-gemara v0.7.0
	github.com/maypok86/otter/v2 v2.3.0
	github.com/oapi-codegen/runtime v1.4.1
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.58.0
	go.opentelemetry.io/collector/component/componenttest v0.152.0
	go.opentelemetry.io/collector/config/confighttp v0.152.0
	go.opentelemetry.io/collector/consumer v1.58.0
	go.opentelemetry.io/collector/pdata v1.58.0
	go.opentelemetry.io/collector/processor v1.58.0
	go.opentelemetry.io/collector/processor/processorhelper v0.152.0
	go.opentelemetry.io/collector/processor/processortest v0.152.0
	go.uber.org/zap v1.28.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/charmbracelet/bubbles v1.0.0 // indirect
	github.com/charmbracelet/bubbletea v1.3.10 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/log v1.0.0 // indirect
	github.com/charmbracelet/x/ansi v0.11.7 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/exp/golden v0.0.0-20260525135217-abeec2b8bf0b // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20260427185012-515ba073c4c1 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fzipp/gocyclo v0.6.0 // indirect
	github.com/getkin/kin-openapi v0.132.0 // indirect
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.9-0.20260124013517-8f8f42cba0de // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.4 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.5.0 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.42.0 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.27 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/speakeasy-api/jsonpath v0.6.0 // indirect
	github.com/speakeasy-api/openapi-overlay v0.10.2 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/unbound-force/gaze v1.5.0 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/client v1.58.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.152.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.58.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.58.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.58.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.58.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.58.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.58.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.58.0 // indirect
	go.opentelemetry.io/collector/confmap v1.58.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.152.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.152.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.152.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.58.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.152.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.58.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.152.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.152.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.152.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.58.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.152.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260615183401-62b3387ff324 // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// OPENTELEMETRY VERSION CONSTRAINT
// ---------------------------------
// OTel Collector packages are pinned to v1.58.0 (stable) and v0.152.0 (experimental) to match
// the requirements of opentelemetry-collector-contrib v0.152.0 (used in beacon-distro/manifest.yaml).
//
// These versions are NOT automatically updated by `task dev:deps:update` — that script explicitly
// excludes OTel packages to prevent version drift. Update OTel versions manually after verifying
// contrib package availability:
//
//   1. Check latest contrib release: https://github.com/open-telemetry/opentelemetry-collector-contrib/releases
//   2. Identify the required OTel version (check a contrib package's go.mod)
//   3. Update truthbeam: cd truthbeam && go get go.opentelemetry.io/collector/component@vX.Y.Z ...
//   4. Run: task version:sync
//   5. Verify: task test && task integration:test
//
// Context: Contrib packages release 1-2 versions behind the main collector packages, causing builds
// to fail if we blindly upgrade to the latest OTel version.
