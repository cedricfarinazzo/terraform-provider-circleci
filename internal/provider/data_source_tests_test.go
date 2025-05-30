package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTestsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccTestsDataSourceConfig("gh/test-org/test-repo", 123),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_tests.test", "project_slug", "gh/test-org/test-repo"),
					resource.TestCheckResourceAttr("data.circleci_tests.test", "job_number", "123"),
					resource.TestCheckResourceAttrSet("data.circleci_tests.test", "tests.#"),
				),
			},
		},
	})
}

func TestAccTestsDataSource_withFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTestsDataSourceConfigWithFilters("gh/test-org/test-repo", 123),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_tests.test", "project_slug", "gh/test-org/test-repo"),
					resource.TestCheckResourceAttr("data.circleci_tests.test", "job_number", "123"),
					resource.TestCheckResourceAttrSet("data.circleci_tests.test", "tests.#"),
				),
			},
		},
	})
}

func testAccTestsDataSourceConfig(projectSlug string, jobNumber int) string {
	return fmt.Sprintf(`
data "circleci_tests" "test" {
  project_slug = "%s"
  job_number   = %d
}
`, projectSlug, jobNumber)
}

func testAccTestsDataSourceConfigWithFilters(projectSlug string, jobNumber int) string {
	return fmt.Sprintf(`
data "circleci_tests" "test" {
  project_slug = "%s"
  job_number   = %d
  
  filter = {
    result = "failure"
  }
}
`, projectSlug, jobNumber)
}
