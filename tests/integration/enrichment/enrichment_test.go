package enrichment_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var _ = Describe("Enrichment Layer", func() {
	Describe("Enrichment Pipeline", func() {
		It("enriches evidence with compliance metadata", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL,
					`{policy_rule_id="github_branch_protection"} | compliance_enrichment_status="Success"`,
				)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})

		It("marks unknown policies as Unmapped", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-unknown.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL,
					`{policy_rule_id="unknown_policy_xyz"} | compliance_enrichment_status="Unmapped"`,
				)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())

			// Verify pipeline is still healthy after processing unknown policy.
			resp2, err := integration.HTTPClient.Get(webhookURL + "/eventreceiver/healthcheck")
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))
		})
	})
})
