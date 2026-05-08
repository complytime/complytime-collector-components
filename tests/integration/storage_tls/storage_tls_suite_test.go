package storage_tls_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/complytime/complybeacon/tests/integration"
)

var (
	webhookURL string
	s3Bucket   string
)

func TestStorageTLS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage TLS Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	s3Bucket = integration.EnvOrDefault("S3_BUCKET", "complybeacon-evidence")

	Expect(integration.CheckStackRunning(webhookURL, "storage-tls")).To(Succeed())
})
