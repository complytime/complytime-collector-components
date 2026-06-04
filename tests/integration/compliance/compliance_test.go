package compliance_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

// complianceEvidence is an OCSF ScanActivity structure matching the existing
// evidence fixture format. The collector's transform/ocsf processor only
// extracts attributes from this format.
type complianceEvidence struct {
	ActivityID   int              `json:"activity_id"`
	ActivityName string           `json:"activity_name"`
	CategoryName string           `json:"category_name"`
	CategoryUID  int              `json:"category_uid"`
	ClassName    string           `json:"class_name"`
	ClassUID     int              `json:"class_uid"`
	Metadata     evidenceMetadata `json:"metadata"`
	Status       string           `json:"status"`
	StatusID     int              `json:"status_id"`
	Severity     string           `json:"severity"`
	SeverityID   int              `json:"severity_id"`
	Policy       evidencePolicy   `json:"policy"`
	Action       string           `json:"action"`
	ActionID     int              `json:"action_id"`
	Scan         evidenceScan     `json:"scan"`
	Time         int64            `json:"time"`
	TypeName     string           `json:"type_name"`
	TypeUID      int              `json:"type_uid"`
}

type evidenceMetadata struct {
	LogProvider string          `json:"log_provider"`
	Product     evidenceProduct `json:"product"`
	UID         string          `json:"uid"`
	Version     string          `json:"version"`
}

type evidenceProduct struct {
	Name       string `json:"name"`
	VendorName string `json:"vendor_name"`
	Version    string `json:"version"`
}

type evidencePolicy struct {
	Data string `json:"data"`
	Desc string `json:"desc"`
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type evidenceScan struct {
	TypeID int `json:"type_id"`
}

// policyData is the JSON-encoded content placed inside Policy.Data.
// The collector's transform/ocsf processor double-parses this field.
type policyData struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Sources     []policyDataSource `json:"sources"`
}

type policyDataSource struct {
	Name   string           `json:"name"`
	Config policyDataConfig `json:"config"`
}

type policyDataConfig struct {
	Include []string `json:"include"`
}

// buildEvidence constructs an OCSF ScanActivity from a control ID and
// requirement ID, matching the format expected by the collector pipeline.
func buildEvidence(controlID, requirementID, catalogID string) ([]byte, error) {
	sources := policyData{
		Name:        "AMPEL Policy",
		Description: "Policy from ampel-branch-protection bundle",
		Sources: []policyDataSource{
			{
				Name:   requirementID,
				Config: policyDataConfig{Include: []string{controlID}},
			},
		},
	}
	sourcesJSON, err := json.Marshal(sources)
	if err != nil {
		return nil, fmt.Errorf("marshaling policy data: %w", err)
	}

	ev := complianceEvidence{
		ActivityID:   0,
		ActivityName: "",
		CategoryName: "Application Activity",
		CategoryUID:  6,
		ClassName:    "Scan Activity",
		ClassUID:     6007,
		Metadata: evidenceMetadata{
			LogProvider: "ampel",
			Product:     evidenceProduct{Name: "ampel", VendorName: "ampel", Version: "v0.1.0"},
			UID:         fmt.Sprintf("ampel-%s-%s", catalogID, controlID),
			Version:     "v0.1.0",
		},
		Status:     "success",
		StatusID:   1,
		Severity:   "unknown",
		SeverityID: 0,
		Policy: evidencePolicy{
			Data: string(sourcesJSON),
			Desc: fmt.Sprintf("Assessment for %s", controlID),
			Name: "AMPEL Policy",
			UID:  controlID,
		},
		Action:   "observed",
		ActionID: 3,
		Scan:     evidenceScan{TypeID: 0},
		Time:     time.Now().UnixMilli(),
		TypeName: "",
		TypeUID:  60070,
	}

	return json.Marshal(ev)
}

var _ = Describe("ComplyTime Policies Integration", func() {
	Describe("Policy Parsing", func() {
		It("loads the ampel-branch-protection catalog with 5 controls", func() {
			Expect(catalog.Metadata.Id).To(Equal("repo-branch-protection"))
			Expect(catalog.Controls).To(HaveLen(5))

			controlIDs := make([]string, len(catalog.Controls))
			for i, c := range catalog.Controls {
				controlIDs[i] = c.Id
			}
			Expect(controlIDs).To(ConsistOf(
				"pull-request-enforcement",
				"approval-requirements",
				"force-push-restriction",
				"admin-bypass-prevention",
				"code-owner-enforcement",
			))
		})

		It("loads the ampel-branch-protection policy with assessment plans", func() {
			Expect(policy.Metadata.Id).To(Equal("ampel-branch-protection-policy"))
			Expect(policy.Adherence.AssessmentPlans).To(HaveLen(5))
		})

		It("links policy assessment plans to catalog requirements", func() {
			// Build a set of requirement IDs from the catalog
			requirementIDs := map[string]bool{}
			for _, ctrl := range catalog.Controls {
				for _, req := range ctrl.AssessmentRequirements {
					requirementIDs[req.Id] = true
				}
			}

			// Every assessment plan should reference a valid requirement
			for _, plan := range policy.Adherence.AssessmentPlans {
				Expect(requirementIDs).To(HaveKey(plan.RequirementId),
					"assessment plan %q references unknown requirement %q", plan.Id, plan.RequirementId)
			}
		})
	})

	Describe("Evidence Pipeline", func() {
		// Build a control-to-requirement map for evidence construction.
		// Each control has one assessment requirement in this catalog.
		controlRequirements := map[string]string{
			"pull-request-enforcement": "require-pull-request",
			"approval-requirements":    "minimum-approvals",
			"force-push-restriction":   "block-force-push",
			"admin-bypass-prevention":  "prevent-admin-bypass",
			"code-owner-enforcement":   "require-code-owner-review",
		}

		for controlID, requirementID := range controlRequirements {
			It(fmt.Sprintf("sends evidence for %s and verifies in Loki", controlID), func() {
				body, err := buildEvidence(controlID, requirementID, catalog.Metadata.Id)
				Expect(err).NotTo(HaveOccurred())

				resp, err := integration.PostEvidenceBytes(webhookURL, body)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				Eventually(func() ([]string, error) {
					return integration.QueryLoki(lokiURL,
						fmt.Sprintf(`{policy_rule_id="%s"}`, controlID),
					)
				}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
			})

			It(fmt.Sprintf("verifies S3 partitioning for %s", controlID), func() {
				body, err := buildEvidence(controlID, requirementID, catalog.Metadata.Id)
				Expect(err).NotTo(HaveOccurred())

				resp, err := integration.PostEvidenceBytes(webhookURL, body)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				Eventually(func() ([]string, error) {
					return integration.ListS3Objects(s3URL, s3Bucket, controlID+"/")
				}, 30*time.Second, 3*time.Second).Should(
					ContainElement(ContainSubstring("evidence_")),
				)
			})
		}
	})

	Describe("Enrichment", func() {
		It("enriches evidence for mapped controls with compliance metadata", func() {
			// Use pull-request-enforcement as the representative control.
			// Its compass response is configured in compass-responses.json.
			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL,
					`{policy_rule_id="pull-request-enforcement"} | compliance_enrichment_status="Success"`,
				)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})
	})
})
