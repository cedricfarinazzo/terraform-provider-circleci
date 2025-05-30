package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRunnerTokenResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRunnerTokenResourceConfig("my-org/test-runner-class", "test-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_runner_token.test", "resource_class", "my-org/test-runner-class"),
					resource.TestCheckResourceAttr("circleci_runner_token.test", "nickname", "test-token"),
					resource.TestCheckResourceAttrSet("circleci_runner_token.test", "id"),
					resource.TestCheckResourceAttrSet("circleci_runner_token.test", "token"),
					resource.TestCheckResourceAttrSet("circleci_runner_token.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "circleci_runner_token.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The token is sensitive and won't be returned on import
				ImportStateVerifyIgnore: []string{"token"},
			},
			// Update testing (should force replacement since tokens are immutable)
			{
				Config: testAccRunnerTokenResourceConfig("my-org/test-runner-class", "test-token-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_runner_token.test", "nickname", "test-token-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRunnerTokenResource_minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunnerTokenResourceConfigMinimal("my-org/minimal-runner-class"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_runner_token.test", "resource_class", "my-org/minimal-runner-class"),
					resource.TestCheckResourceAttrSet("circleci_runner_token.test", "id"),
					resource.TestCheckResourceAttrSet("circleci_runner_token.test", "token"),
				),
			},
		},
	})
}

func testAccRunnerTokenResourceConfig(resourceClass, nickname string) string {
	return `
resource "circleci_runner_token" "test" {
  resource_class = "` + resourceClass + `"
  nickname       = "` + nickname + `"
}
`
}

func testAccRunnerTokenResourceConfigMinimal(resourceClass string) string {
	return `
resource "circleci_runner_token" "test" {
  resource_class = "` + resourceClass + `"
}
`
}
