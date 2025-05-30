package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContextResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContextResourceConfig("test-context"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_context.test", "name", "test-context"),
					resource.TestCheckResourceAttrSet("circleci_context.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "circleci_context.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccContextResourceConfig("test-context-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("circleci_context.test", "name", "test-context-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccContextResourceConfig(name string) string {
	return `
resource "circleci_context" "test" {
  name = "` + name + `"
  owner = {
    id   = "test-org-id"
    slug = "github"
    type = "organization"
  }
}
`
}
