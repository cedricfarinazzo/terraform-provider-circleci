terraform {
  required_providers {
    circleci = {
      source  = "your-org/circleci"
      version = "~> 1.0"
    }
  }
}

provider "circleci" {
  api_token = var.circleci_token
}

variable "circleci_token" {
  description = "CircleCI API token"
  type        = string
  sensitive   = true
}

variable "project_slug" {
  description = "Project slug (e.g., 'gh/my-org/my-repo')"
  type        = string
}

variable "workflow_id" {
  description = "Workflow ID to analyze"
  type        = string
}

variable "job_number" {
  description = "Job number to analyze"
  type        = number
}

# Get artifacts from a specific job
data "circleci_artifacts" "build_artifacts" {
  project_slug = var.project_slug
  job_number   = var.job_number
  
  filter = {
    path_pattern = "*.jar"  # Only get JAR files
    node_index   = 0        # From first node
  }
}

# Get test results from a job
data "circleci_tests" "test_results" {
  project_slug = var.project_slug
  job_number   = var.job_number
  
  filter = {
    result = "failure"  # Only failed tests
  }
}

# Get all jobs in a workflow
data "circleci_jobs" "workflow_jobs" {
  workflow_id = var.workflow_id
}

# Get only failed jobs from the workflow
data "circleci_jobs" "failed_jobs" {
  workflow_id = var.workflow_id
  
  filter = {
    status = "failed"
  }
}

# Get approval jobs from the workflow
data "circleci_jobs" "approval_jobs" {
  workflow_id = var.workflow_id
  
  filter = {
    type = "approval"
  }
}

# Outputs for analysis
output "artifact_count" {
  description = "Number of JAR artifacts found"
  value       = length(data.circleci_artifacts.build_artifacts.artifacts)
}

output "artifact_urls" {
  description = "URLs of all JAR artifacts"
  value       = [for artifact in data.circleci_artifacts.build_artifacts.artifacts : artifact.url]
}

output "failed_test_count" {
  description = "Number of failed tests"
  value       = length(data.circleci_tests.test_results.tests)
}

output "failed_test_names" {
  description = "Names of failed tests"
  value       = [for test in data.circleci_tests.test_results.tests : test.name]
}

output "total_test_time" {
  description = "Total execution time of failed tests (seconds)"
  value = sum([
    for test in data.circleci_tests.test_results.tests : test.run_time_seconds
  ])
}

output "workflow_job_statuses" {
  description = "Status of all jobs in the workflow"
  value = {
    for job in data.circleci_jobs.workflow_jobs.jobs : job.name => job.status
  }
}

output "failed_job_names" {
  description = "Names of failed jobs"
  value = [for job in data.circleci_jobs.failed_jobs.jobs : job.name]
}

output "pending_approvals" {
  description = "Jobs waiting for approval"
  value = [
    for job in data.circleci_jobs.approval_jobs.jobs : {
      name      = job.name
      job_number = job.job_number
      approval_request_id = job.approval_request_id
    }
    if job.status == "on_hold"
  ]
}

# Example of creating a summary report
locals {
  workflow_summary = {
    total_jobs        = length(data.circleci_jobs.workflow_jobs.jobs)
    failed_jobs       = length(data.circleci_jobs.failed_jobs.jobs)
    pending_approvals = length([for job in data.circleci_jobs.approval_jobs.jobs : job if job.status == "on_hold"])
    success_rate      = (length(data.circleci_jobs.workflow_jobs.jobs) - length(data.circleci_jobs.failed_jobs.jobs)) / length(data.circleci_jobs.workflow_jobs.jobs) * 100
  }
}

output "workflow_summary" {
  description = "Summary of workflow execution"
  value = local.workflow_summary
}
