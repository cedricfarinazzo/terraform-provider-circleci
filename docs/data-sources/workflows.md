# circleci_workflows

Retrieves information about workflows within a specific CircleCI pipeline.

## Example Usage

```hcl
data "circleci_workflows" "pipeline_workflows" {
  pipeline_id = "5034460f-c7c4-4c43-9457-de07e2029e7b"
}

# Output workflow statuses
output "workflow_statuses" {
  value = {
    for workflow in data.circleci_workflows.pipeline_workflows.workflows :
    workflow.name => workflow.status
  }
}

# Find failed workflows
locals {
  failed_workflows = [
    for workflow in data.circleci_workflows.pipeline_workflows.workflows :
    workflow if workflow.status == "failed"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `pipeline_id` - (Required) The pipeline ID for which to retrieve workflows.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `workflows` - A list of workflows in the pipeline. Each workflow has the following attributes:
  * `id` - The unique identifier of the workflow.
  * `name` - The name of the workflow.
  * `pipeline_id` - The pipeline ID that contains this workflow.
  * `project_slug` - The project slug for the workflow.
  * `status` - The current status of the workflow.
  * `started_by` - The user who started the workflow.
  * `created_at` - The date and time the workflow was created.
  * `stopped_at` - The date and time the workflow stopped (if applicable).
  * `tag` - The tag associated with the workflow (if applicable).

## Notes

* This data source is useful for monitoring pipeline execution and identifying workflow statuses.
* Workflows are returned in the order they were created within the pipeline.
* The `stopped_at` field is only populated for completed workflows.
