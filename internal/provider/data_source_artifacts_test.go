package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArtifactsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccArtifactsDataSourceConfig("gh/test-org/test-repo", 123),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_artifacts.test", "project_slug", "gh/test-org/test-repo"),
					resource.TestCheckResourceAttr("data.circleci_artifacts.test", "job_number", "123"),
					resource.TestCheckResourceAttrSet("data.circleci_artifacts.test", "artifacts.#"),
				),
			},
		},
	})
}

func TestAccArtifactsDataSource_withFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArtifactsDataSourceConfigWithFilters("gh/test-org/test-repo", 123),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_artifacts.test", "project_slug", "gh/test-org/test-repo"),
					resource.TestCheckResourceAttr("data.circleci_artifacts.test", "job_number", "123"),
					resource.TestCheckResourceAttrSet("data.circleci_artifacts.test", "artifacts.#"),
				),
			},
		},
	})
}

func testAccArtifactsDataSourceConfig(projectSlug string, jobNumber int) string {
	return fmt.Sprintf(`
data "circleci_artifacts" "test" {
  project_slug = "%s"
  job_number   = %d
}
`, projectSlug, jobNumber)
}

func testAccArtifactsDataSourceConfigWithFilters(projectSlug string, jobNumber int) string {
	return fmt.Sprintf(`
data "circleci_artifacts" "test" {
  project_slug = "%s"
  job_number   = %d
  
  filter = {
    path_pattern = "*.jar"
    node_index   = 0
  }
}
`, projectSlug, jobNumber)
}
