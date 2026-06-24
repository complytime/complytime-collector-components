package auth_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Auth Layer", func() {
	Describe("Inbound OIDC Authentication", func() {
		It("rejects unauthenticated OTLP requests", func() {
			err := integration.PostEvidenceOTLP(otlpAddr, "../fixtures/evidence-fail.json", "")
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "expected gRPC status error")
			Expect(st.Code()).To(Equal(codes.Unauthenticated))
		})

		It("rejects OTLP requests with an invalid token", func() {
			err := integration.PostEvidenceOTLP(otlpAddr, "../fixtures/evidence-fail.json", "invalid-garbage-token")
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "expected gRPC status error")
			Expect(st.Code()).To(Equal(codes.Unauthenticated))
		})

		It("rejects an expired OIDC token", func() {
			token, err := integration.MintDexToken(dexURL, "beacon-collector", "test@complybeacon.dev", "testpassword")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(7 * time.Second)

			err = integration.PostEvidenceOTLP(otlpAddr, "../fixtures/evidence-fail.json", token)
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "expected gRPC status error")
			Expect(st.Code()).To(Equal(codes.Unauthenticated))
		})

		It("rejects a valid token with the wrong audience", func() {
			token, err := integration.MintDexToken(dexURL, "wrong-audience", "test@complybeacon.dev", "testpassword")
			Expect(err).NotTo(HaveOccurred())

			err = integration.PostEvidenceOTLP(otlpAddr, "../fixtures/evidence-fail.json", token)
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "expected gRPC status error")
			Expect(st.Code()).To(Equal(codes.Unauthenticated))
		})

		It("accepts OTLP requests with a valid OIDC token", func() {
			token, err := integration.MintDexToken(dexURL, "beacon-collector", "test@complybeacon.dev", "testpassword")
			Expect(err).NotTo(HaveOccurred(), "failed to mint token from Dex")
			Expect(token).NotTo(BeEmpty())

			err = integration.PostEvidenceOTLP(otlpAddr, "../fixtures/evidence-fail.json", token)
			Expect(err).NotTo(HaveOccurred(), "authenticated OTLP request should succeed")

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL, `{policy_rule_id="github_branch_protection"}`)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})
	})

	Describe("Webhook Receiver", func() {
		It("remains accessible without authentication", func() {
			resp, err := integration.PostEvidence(webhookURL, "../fixtures/evidence-pass.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Eventually(func() ([]string, error) {
				return integration.QueryLoki(lokiURL, `{policy_rule_id="code_review"}`)
			}, 30*time.Second, 3*time.Second).ShouldNot(BeEmpty())
		})
	})
})
