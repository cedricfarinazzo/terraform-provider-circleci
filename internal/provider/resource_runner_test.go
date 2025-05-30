package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRunnerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRunnerResourceConfig("my-org/test-runner-class", "test-runner", "Test runner for CI"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_runner.test", "resource_class", "my-org/test-runner-class"),
					resource.TestCheckResourceAttr("circleci_runner.test", "name", "test-runner"),
					resource.TestCheckResourceAttr("circleci_runner.test", "description", "Test runner for CI"),
					resource.TestCheckResourceAttrSet("circleci_runner.test", "id"),
					resource.TestCheckResourceAttrSet("circleci_runner.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "circleci_runner.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccRunnerResourceConfig("my-org/test-runner-class", "test-runner-updated", "Updated test runner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_runner.test", "name", "test-runner-updated"),
					resource.TestCheckResourceAttr("circleci_runner.test", "description", "Updated test runner"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRunnerResource_minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunnerResourceConfigMinimal("my-org/minimal-runner-class", "minimal-runner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_runner.test", "resource_class", "my-org/minimal-runner-class"),
					resource.TestCheckResourceAttr("circleci_runner.test", "name", "minimal-runner"),
					resource.TestCheckResourceAttrSet("circleci_runner.test", "id"),
				),
			},
		},
	})
}

func testAccRunnerResourceConfig(resourceClass, name, description string) string {
	return `
resource "circleci_runner" "test" {
  resource_class = "` + resourceClass + `"
  name           = "` + name + `"
  description    = "` + description + `"
}
`
}

func testAccRunnerResourceConfigMinimal(resourceClass, name string) string {
	return `
resource "circleci_runner" "test" {
  resource_class = "` + resourceClass + `"
  name           = "` + name + `"
}
`
}
