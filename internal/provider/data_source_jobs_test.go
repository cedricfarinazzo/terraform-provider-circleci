package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJobsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccJobsDataSourceConfig("5034460f-c7c4-4c43-9457-de07e2029e7b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_jobs.test", "workflow_id", "5034460f-c7c4-4c43-9457-de07e2029e7b"),
					resource.TestCheckResourceAttrSet("data.circleci_jobs.test", "jobs.#"),
				),
			},
		},
	})
}

func TestAccJobsDataSource_withFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobsDataSourceConfigWithFilters("5034460f-c7c4-4c43-9457-de07e2029e7b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_jobs.test", "workflow_id", "5034460f-c7c4-4c43-9457-de07e2029e7b"),
					resource.TestCheckResourceAttrSet("data.circleci_jobs.test", "jobs.#"),
				),
			},
		},
	})
}

func testAccJobsDataSourceConfig(workflowId string) string {
	return `
data "circleci_jobs" "test" {
  workflow_id = "` + workflowId + `"
}
`
}

func testAccJobsDataSourceConfigWithFilters(workflowId string) string {
	return `
data "circleci_jobs" "test" {
  workflow_id = "` + workflowId + `"
  
  filter = {
    status = "failed"
    type   = "approval"
  }
}
`
}
