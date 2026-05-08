package storage_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var _ = Describe("Storage Layer", func() {
	Describe("S3 Evidence Export", func() {
		It("exports evidence objects to S3", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.ListS3Objects(s3URL, s3Bucket, "github_branch_protection/")
			}, 30*time.Second, 3*time.Second).Should(
				ContainElement(ContainSubstring("evidence_")),
			)
		})

		It("partitions by policy_rule_id", func() {
			resp1, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp1.Body.Close()
			Expect(resp1.StatusCode).To(Equal(http.StatusOK))

			resp2, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-pass.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.ListS3Objects(s3URL, s3Bucket, "github_branch_protection/")
			}, 30*time.Second, 3*time.Second).Should(
				ContainElement(ContainSubstring("evidence_")),
			)

			Eventually(func() ([]string, error) {
				return integration.ListS3Objects(s3URL, s3Bucket, "code_review/")
			}, 30*time.Second, 3*time.Second).Should(
				ContainElement(ContainSubstring("evidence_")),
			)
		})
	})
})
