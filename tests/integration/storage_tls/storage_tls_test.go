package storage_tls_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var _ = Describe("Storage TLS Layer", func() {
	Describe("TLS S3 Export", func() {
		It("exports evidence to S3 over TLS", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() (string, error) {
				return integration.ExecInContainer("rustfs-init-tls",
					"rc", "ls", "--insecure",
					"local/"+s3Bucket+"/github_branch_protection/",
				)
			}, 30*time.Second, 3*time.Second).Should(ContainSubstring("evidence_"))
		})

		It("partitions by policy_rule_id over TLS", func() {
			resp1, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-fail.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp1.Body.Close()
			Expect(resp1.StatusCode).To(Equal(http.StatusOK))

			resp2, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-pass.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() (string, error) {
				return integration.ExecInContainer("rustfs-init-tls",
					"rc", "ls", "--insecure",
					"local/"+s3Bucket+"/github_branch_protection/",
				)
			}, 30*time.Second, 3*time.Second).Should(ContainSubstring("evidence_"))

			Eventually(func() (string, error) {
				return integration.ExecInContainer("rustfs-init-tls",
					"rc", "ls", "--insecure",
					"local/"+s3Bucket+"/code_review/",
				)
			}, 30*time.Second, 3*time.Second).Should(ContainSubstring("evidence_"))
		})
	})
})
