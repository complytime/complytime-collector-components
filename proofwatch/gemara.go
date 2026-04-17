package proofwatch

import (
	"encoding/json"
	"time"

	gemara "github.com/gemaraproj/go-gemara"
	"go.opentelemetry.io/otel/attribute"
)

var _ Evidence = (*GemaraEvidence)(nil)

// GemaraEvidence represents evidence data from the Gemara compliance assessment framework.
type GemaraEvidence struct {
	gemara.Metadata
	gemara.AssessmentLog
}

func (g GemaraEvidence) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g GemaraEvidence) Attributes() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(POLICY_ENGINE_NAME, g.Author.Name),
		attribute.String(COMPLIANCE_CONTROL_ID, g.Requirement.EntryId),
		attribute.String(COMPLIANCE_CONTROL_CATALOG_ID, g.Requirement.ReferenceId),
		attribute.String(POLICY_EVALUATION_RESULT, g.Result.String()),
		attribute.String(COMPLIANCE_ASSESSMENT_ID, g.Id),
	}

	if g.Plan != nil {
		attrs = append(attrs, attribute.String(POLICY_RULE_ID, g.Plan.EntryId))
	}

	if g.Message != "" {
		attrs = append(attrs, attribute.String(POLICY_EVALUATION_MESSAGE, g.Message))
	}

	if g.Recommendation != "" {
		attrs = append(attrs, attribute.String(COMPLIANCE_REMEDIATION_DESCRIPTION, g.Recommendation))
	}

	return attrs
}

func (g GemaraEvidence) Timestamp() time.Time {
	timestamp, err := time.Parse(time.RFC3339, string(g.End))
	if err != nil {
		return time.Now()
	}
	return timestamp
}
