package enrichment_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var (
	webhookURL string
	lokiURL    string
)

func TestEnrichment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Enrichment Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	lokiURL = integration.EnvOrDefault("LOKI_URL", "http://localhost:3100")

	Expect(integration.CheckStackRunning(webhookURL, "enrichment")).To(Succeed())
})
