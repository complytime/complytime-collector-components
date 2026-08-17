package proofwatch

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	olog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/complytime/complybeacon/proofwatch/internal/metrics"
)

const (
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/complytime/complybeacon/proofwatch"
)

type ProofWatch struct {
	logger        olog.Logger
	tracer        trace.Tracer
	observer      *metrics.EvidenceObserver
	levelSeverity olog.Severity
}

// NewProofWatch creates a new ProofWatch instance with OpenTelemetry logging.
func NewProofWatch(opts ...OptionFunc) (*ProofWatch, error) {
	cfg := config{
		MeterProvider:  otel.GetMeterProvider(),
		LoggerProvider: global.GetLoggerProvider(),
		TracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	meter := cfg.MeterProvider.Meter(ScopeName, metric.WithInstrumentationVersion(Version()))
	observer, err := metrics.NewEvidenceObserver(meter)
	if err != nil {
		return nil, err
	}
	return &ProofWatch{
		logger:   cfg.LoggerProvider.Logger(ScopeName, olog.WithInstrumentationVersion(Version())),
		tracer:   cfg.TracerProvider.Tracer(ScopeName, trace.WithInstrumentationVersion(Version())),
		observer: observer,
		// Default severity
		levelSeverity: olog.SeverityInfo,
	}, nil
}

// Log logs a policy event using OpenTelemetry's log API.
func (w *ProofWatch) Log(ctx context.Context, evidence Evidence) error {
	return w.LogWithSeverity(ctx, evidence, w.levelSeverity)
}

// LogWithSeverity logs a policy event using OpenTelemetry's log API with a given severity level
func (w *ProofWatch) LogWithSeverity(ctx context.Context, evidence Evidence, severity olog.Severity) error {

	ctx, span := w.tracer.Start(ctx, "evidence.log_evidence")
	defer span.End()

	attrs := evidence.Attributes()

	jsonData, err := evidence.ToJSON()
	if err != nil {
		return err
	}

	record := olog.Record{}
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetObservedTimestamp(time.Now())
	// Set event time
	record.SetTimestamp(evidence.Timestamp())
	record.AddAttributes(attrs...)
	record.SetBody(attribute.StringValue(string(jsonData))) // Retains the original body for flexibility.

	span.AddEvent("evidence.logged", trace.WithAttributes(attrs...), trace.WithTimestamp(time.Now()))

	w.logger.Emit(ctx, record)

	w.observer.Processed(ctx, attrs...)

	return nil
}

// Version is the current release version of Proofwatch
func Version() string {
	return "0.1.0"
}
