# Data Source: circleci_workflow

Retrieve information about a specific CircleCI workflow.

## Example Usage

```hcl
data "circleci_workflow" "build_workflow" {
  id = "5034460f-c7c4-4c43-9457-de07e2029e7b"
}

output "workflow_status" {
  value = data.circleci_workflow.build_workflow.status
}

output "workflow_created_at" {
  value = data.circleci_workflow.build_workflow.created_at
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The unique ID of the workflow to retrieve.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name of the workflow.
* `project_slug` - The project slug in the form `vcs-slug/org-name/repo-name`.
* `pipeline_id` - The ID of the pipeline that contains this workflow.
* `pipeline_number` - The number of the pipeline that contains this workflow.
* `status` - The status of the workflow (e.g., "success", "failed", "running", "canceled").
* `started_by` - The ID of the user who started the workflow.
* `created_at` - The timestamp when the workflow was created.
* `stopped_at` - The timestamp when the workflow stopped (if applicable).
* `tag` - The tag associated with the workflow (if any).

## Import

Workflows can be referenced using their ID:

```hcl
data "circleci_workflow" "example" {
  id = "5034460f-c7c4-4c43-9457-de07e2029e7b"
}
```
