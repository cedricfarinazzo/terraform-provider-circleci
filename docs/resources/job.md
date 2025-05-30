# circleci_job

Manages a CircleCI job. This resource allows you to interact with existing jobs, including canceling, approving, or rerunning them.

## Example Usage

```hcl
# Monitor a specific job
resource "circleci_job" "build_job" {
  project_slug = "gh/myorg/myrepo"
  job_number   = 123
}

# Cancel a running job
resource "circleci_job" "cancel_job" {
  project_slug = "gh/myorg/myrepo"
  job_number   = 456
  action       = "cancel"
}

# Approve a job waiting for approval
resource "circleci_job" "approve_job" {
  project_slug = "gh/myorg/myrepo"
  job_number   = 789
  action       = "approve"
  approval_id  = "approval-request-id"
}

# Rerun a failed job
resource "circleci_job" "rerun_job" {
  project_slug = "gh/myorg/myrepo"
  job_number   = 101
  action       = "rerun"
}
```

## Argument Reference

The following arguments are supported:

* `project_slug` - (Required) The project slug in the form `vcs-slug/org-name/repo-name`. Changing this forces a new resource to be created.
* `job_number` - (Required) The job number. Changing this forces a new resource to be created.
* `action` - (Optional) Action to perform on the job. Valid values are `cancel`, `approve`, and `rerun`.
* `approval_id` - (Optional) The approval request ID, required when using the `approve` action.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the job.
* `name` - The name of the job.
* `status` - The current status of the job.
* `started_at` - The date and time the job started.
* `stopped_at` - The date and time the job stopped (if applicable).
* `approval_type` - The type of approval required (for approval jobs).

## Import

Jobs can be imported using the project slug and job number separated by a colon:

```
terraform import circleci_job.example gh/myorg/myrepo:123
```

## Notes

* Job resources are primarily read-only. The main purpose is to monitor job status and perform actions.
* Deleting a job resource only removes it from Terraform state; the actual job continues to exist in CircleCI.
* Actions (`cancel`, `approve`, `rerun`) are performed when the resource is created or updated.
* The `approval_id` can be obtained from the CircleCI API or web interface for jobs requiring approval.
