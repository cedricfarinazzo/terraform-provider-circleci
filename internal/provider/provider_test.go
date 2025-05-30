package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the CircleCI client is properly configured.
	providerConfig = `
provider "circleci" {
  api_token = "test-token"
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"circleci": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if v := os.Getenv("CIRCLECI_TOKEN"); v == "" {
		t.Fatal("CIRCLECI_TOKEN must be set for acceptance tests")
	}
}

func TestProvider(t *testing.T) {
	provider := New("test")()

	// Test that the provider can be created without error
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}
}
