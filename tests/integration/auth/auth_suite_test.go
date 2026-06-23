package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var (
	webhookURL string
	lokiURL    string
	otlpAddr   string
	dexURL     string
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	lokiURL = integration.EnvOrDefault("LOKI_URL", "http://localhost:3100")
	otlpAddr = integration.EnvOrDefault("OTLP_ADDR", "localhost:14317")
	dexURL = integration.EnvOrDefault("DEX_URL", "http://localhost:15556")

	Expect(integration.CheckStackRunning(webhookURL, "auth")).To(Succeed())
})
