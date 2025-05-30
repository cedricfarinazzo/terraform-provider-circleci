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

variable "organization_id" {
  description = "Organization ID"
  type        = string
}

# Get project information
data "circleci_project" "example" {
  slug = var.project_slug
}

# Get organization information
data "circleci_organization" "main" {
  id = var.organization_id
}

# Get organization policies
data "circleci_policies" "org_policies" {
  org_id = data.circleci_organization.main.id
}

# Get insights for the project
data "circleci_insight" "project_metrics" {
  project_slug = var.project_slug
  workflow     = "main"
  branch       = "main"
}

# Get context information
data "circleci_context" "shared" {
  name = "shared-context"
  owner = {
    id   = data.circleci_organization.main.id
    slug = data.circleci_organization.main.slug
    type = "organization"
  }
}

# Outputs for analysis
output "project_info" {
  description = "Project information"
  value = {
    name         = data.circleci_project.example.name
    organization = data.circleci_project.example.organization
    vcs_url      = data.circleci_project.example.vcs_url
    vcs_type     = data.circleci_project.example.vcs_type
  }
}

output "organization_info" {
  description = "Organization information"
  value = {
    name = data.circleci_organization.main.name
    slug = data.circleci_organization.main.slug
  }
}

output "policy_count" {
  description = "Number of policies in organization"
  value       = length(data.circleci_policies.org_policies.policies)
}

output "project_metrics" {
  description = "Project performance metrics"
  value = {
    success_rate    = data.circleci_insight.project_metrics.metrics.success_rate
    mean_duration   = data.circleci_insight.project_metrics.metrics.mean_duration_sec
    total_runs      = data.circleci_insight.project_metrics.metrics.total_runs
    throughput      = data.circleci_insight.project_metrics.metrics.throughput
  }
}
