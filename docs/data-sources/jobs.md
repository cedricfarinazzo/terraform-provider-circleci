# circleci_jobs

Retrieves information about jobs within a specific CircleCI workflow.

## Example Usage

```hcl
data "circleci_jobs" "workflow_jobs" {
  workflow_id = "5034460f-c7c4-4c43-9457-de07e2029e7b"
}

# Output job statuses
output "job_statuses" {
  value = {
    for job in data.circleci_jobs.workflow_jobs.jobs :
    job.name => job.status
  }
}

# Find failed jobs
locals {
  failed_jobs = [
    for job in data.circleci_jobs.workflow_jobs.jobs :
    job if job.status == "failed"
  ]
}

# Find jobs requiring approval
locals {
  approval_jobs = [
    for job in data.circleci_jobs.workflow_jobs.jobs :
    job if job.approval_type != ""
  ]
}
```

## Argument Reference

The following arguments are supported:

* `workflow_id` - (Required) The workflow ID for which to retrieve jobs.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `jobs` - A list of jobs in the workflow. Each job has the following attributes:
  * `id` - The unique identifier of the job.
  * `name` - The name of the job.
  * `project_slug` - The project slug for the job.
  * `job_number` - The job number.
  * `status` - The current status of the job.
  * `started_at` - The date and time the job started.
  * `stopped_at` - The date and time the job stopped.
  * `approval_type` - The type of approval required (for approval jobs).
  * `approval_request_id` - The approval request ID (for approval jobs).

## Notes

* This data source is useful for monitoring workflow execution and identifying job dependencies.
* Jobs are returned in the order they were created within the workflow.
* The `stopped_at` field is only populated for completed jobs.
* Approval jobs have an `approval_type` and `approval_request_id` for manual approval workflows.
