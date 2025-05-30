# circleci_artifacts

Retrieves artifacts from a CircleCI job.

## Example Usage

```hcl
data "circleci_artifacts" "build_artifacts" {
  project_slug = "gh/myorg/myrepo"
  job_number   = 123
}

# Output artifact URLs
output "artifact_urls" {
  value = [
    for artifact in data.circleci_artifacts.build_artifacts.artifacts :
    artifact.url
  ]
}

# Find specific artifacts
locals {
  test_reports = [
    for artifact in data.circleci_artifacts.build_artifacts.artifacts :
    artifact if can(regex("test-results", artifact.path))
  ]
}
```

## Argument Reference

The following arguments are supported:

* `project_slug` - (Required) Project slug in the form `vcs-slug/org-name/repo-name`.
* `job_number` - (Required) The job number for which to retrieve artifacts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `artifacts` - A list of artifacts from the job. Each artifact has the following attributes:
  * `path` - The path to the artifact file.
  * `node_index` - The node index for the artifact.
  * `url` - The download URL for the artifact.
  * `pretty_path` - The pretty-formatted path to the artifact.

## Notes

* This data source is useful for retrieving build artifacts such as test reports, coverage data, and deployment packages.
* Artifact URLs are temporary and may expire after a certain period.
* The `node_index` indicates which parallel execution node created the artifact.
