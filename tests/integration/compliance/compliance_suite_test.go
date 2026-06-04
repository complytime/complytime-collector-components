package compliance_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"oras.land/oras-go/v2/content/oci"

	"github.com/gemaraproj/go-gemara"
	"github.com/gemaraproj/go-gemara/bundle"
	// Gemara types implement goccy/go-yaml's BytesUnmarshaler interface
	// (UnmarshalYAML([]byte) error), not the yaml.v3 Unmarshaler interface
	// (UnmarshalYAML(*Node) error). Using yaml.v3 here silently falls through
	// to direct type assignment and fails at runtime on enum fields.
	// Tracked upstream: https://github.com/gemaraproj/go-gemara/issues/76
	"github.com/goccy/go-yaml"

	"github.com/complytime/complybeacon/tests/integration"
)

var (
	webhookURL string
	lokiURL    string
	s3URL      string
	s3Bucket   string

	catalog gemara.ControlCatalog
	policy  gemara.Policy
)

func TestCompliance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compliance Layer Suite")
}

var _ = BeforeSuite(func() {
	webhookURL = integration.EnvOrDefault("WEBHOOK_URL", "http://localhost:8088")
	lokiURL = integration.EnvOrDefault("LOKI_URL", "http://localhost:3100")
	s3URL = integration.EnvOrDefault("S3_URL", "http://localhost:9000")
	s3Bucket = integration.EnvOrDefault("S3_BUCKET", "complybeacon-evidence")

	Expect(integration.CheckStackRunning(webhookURL, "compliance")).To(Succeed())

	// Load the ampel-branch-protection bundle from the OCI Layout store.
	// complyctl stores pulled bundles at ~/.complytime/policies/{namespace}/{repo}/
	home, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())

	storePath := filepath.Join(home, ".complytime", "policies", "complytime", "policies-ampel-branch-protection")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		Skip("policy artifacts not available — complyctl pull failed or was skipped")
	}

	store, err := oci.New(storePath)
	Expect(err).NotTo(HaveOccurred(), "failed to open OCI Layout store at %s", storePath)

	ctx := context.Background()
	b, err := bundle.Unpack(ctx, store, "v0.1.0")
	Expect(err).NotTo(HaveOccurred(), "failed to unpack ampel-branch-protection bundle")

	// Imports[0] is the catalog (role=import), Files[0] is the policy (role=artifact)
	Expect(b.Imports).NotTo(BeEmpty(), "bundle has no imports (expected catalog)")
	Expect(b.Files).NotTo(BeEmpty(), "bundle has no files (expected policy)")

	err = yaml.Unmarshal(b.Imports[0].Data, &catalog)
	Expect(err).NotTo(HaveOccurred(), "failed to unmarshal catalog YAML")

	err = yaml.Unmarshal(b.Files[0].Data, &policy)
	Expect(err).NotTo(HaveOccurred(), "failed to unmarshal policy YAML")
})
