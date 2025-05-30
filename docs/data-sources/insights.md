# Data Source: circleci_insights

Retrieve workflow metrics and performance data for a specific project and workflow.

## Example Usage

```hcl
data "circleci_insights" "metrics" {
  project_slug = "gh/my-org/my-repo"
  workflow     = "build-and-test"
  branch       = "main"
}

output "success_rate" {
  value = data.circleci_insights.metrics.metrics.success_rate
}

output "mean_duration" {
  value = data.circleci_insights.metrics.metrics.mean_duration_sec
}
```

## Example Usage with Date Range

```hcl
data "circleci_insights" "monthly_metrics" {
  project_slug = "gh/my-org/my-repo"
  workflow     = "build-and-test"
  branch       = "main"
  
  reporting_window = "last-30-days"
}
```

## Argument Reference

The following arguments are supported:

* `project_slug` - (Required) Project slug in the form `vcs-slug/org-name/repo-name`.
* `workflow` - (Required) The name of the workflow to get insights for.
* `branch` - (Optional) The branch to get insights for. Defaults to all branches.
* `reporting_window` - (Optional) The reporting window for insights. Options: "last-7-days", "last-30-days", "last-60-days", "last-90-days". Defaults to "last-30-days".

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `metrics` - A block containing workflow metrics. Structure is documented below.

The `metrics` block contains:

* `success_rate` - The success rate as a percentage (0-100).
* `total_runs` - The total number of workflow runs.
* `failed_runs` - The number of failed workflow runs.
* `successful_runs` - The number of successful workflow runs.
* `throughput` - The average number of runs per day.
* `mean_duration_sec` - The mean duration of workflow runs in seconds.
* `median_duration_sec` - The median duration of workflow runs in seconds.
* `p95_duration_sec` - The 95th percentile duration of workflow runs in seconds.
* `total_credits_used` - The total credits consumed by workflow runs.

## Import

Insights data sources don't support import as they represent computed metrics rather than stored resources.
