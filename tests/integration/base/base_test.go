package base_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var _ = Describe("Base Layer", func() {
	Describe("Webhook Receiver", func() {
		It("responds to healthcheck", func() {
			resp, err := integration.HTTPClient.Get(webhookURL + "/eventreceiver/healthcheck")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("OCSF Transform Pipeline", func() {
		It("transforms failure evidence to Loki", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL, `{policy_rule_id="github_branch_protection"}`)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})

		It("transforms success evidence to Loki", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-pass.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL, `{policy_rule_id="code_review"}`)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})
	})

	Describe("Resilience", func() {
		It("survives malformed evidence without disruption", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-malformed.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(SatisfyAny(
				Equal(http.StatusOK),
				Equal(http.StatusBadRequest),
			))

			Eventually(func() (int, error) {
				r, err := integration.HTTPClient.Get(webhookURL + "/eventreceiver/healthcheck")
				if err != nil {
					return 0, err
				}
				defer r.Body.Close()
				return r.StatusCode, nil
			}, 6*time.Second, 2*time.Second).Should(Equal(http.StatusOK))

			resp2, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL, `{policy_rule_id="github_branch_protection"}`)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})
	})
})
