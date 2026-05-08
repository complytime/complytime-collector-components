package storage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var (
	webhookURL string
	lokiURL    string
	s3URL      string
	s3Bucket   string
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	lokiURL = integration.EnvOrDefault("LOKI_URL", "http://localhost:3100")
	s3URL = integration.EnvOrDefault("S3_URL", "http://localhost:9000")
	s3Bucket = integration.EnvOrDefault("S3_BUCKET", "complybeacon-evidence")

	Expect(integration.CheckStackRunning(webhookURL, "storage")).To(Succeed())
})
