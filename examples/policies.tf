# CircleCI Policy Examples

terraform {
  required_providers {
    circleci = {
      source = "your-org/circleci"
    }
  }
}

provider "circleci" {
  # Configure with environment variable CIRCLECI_TOKEN
}

# Data source for organization
data "circleci_organization" "main" {
  slug = "github"
  name = "your-org"
}

# Security scanning policy
resource "circleci_policy" "security_scanning" {
  name        = "Security Scanning Requirements"
  description = "Enforce security scanning in all workflows"
  org_id      = data.circleci_organization.main.id
  enabled     = true

  content = <<EOT
package org

import rego.v1

# Deny workflows that don't include security scanning
deny contains msg if {
    input.config.workflows[workflow_name].jobs[_].name
    not has_security_scan(workflow_name)
    msg := sprintf("Workflow '%s' must include security scanning", [workflow_name])
}

has_security_scan(workflow_name) if {
    input.config.workflows[workflow_name].jobs[_].name == "security-scan"
}

# Require specific security tools
deny contains msg if {
    input.config.jobs["security-scan"]
    not uses_required_tools
    msg := "Security scan must use approved tools (bandit, safety, semgrep)"
}

uses_required_tools if {
    job := input.config.jobs["security-scan"]
    job.docker[_].image
    contains(job.docker[_].image, "security-tools")
}
EOT
}

# Branch protection policy
resource "circleci_policy" "branch_protection" {
  name        = "Branch Protection"
  description = "Enforce branch protection rules"
  org_id      = data.circleci_organization.main.id
  enabled     = true

  content = <<EOT
package org

import rego.v1

# Require pull request builds for main branch
deny contains msg if {
    input.build.branch == "main"
    not input.build.pull_request
    msg := "Direct pushes to main branch are not allowed"
}

# Require approval for production deployments
deny contains msg if {
    input.config.workflows[_].jobs[job_name].name == "deploy-production"
    not input.config.jobs[job_name].type == "approval"
    msg := "Production deployments must require manual approval"
}
EOT
}

# Resource constraints policy
resource "circleci_policy" "resource_limits" {
  name        = "Resource Constraints"
  description = "Enforce resource usage limits"
  org_id      = data.circleci_organization.main.id
  enabled     = true

  content = <<EOT
package org

import rego.v1

# Limit resource class usage
deny contains msg if {
    job := input.config.jobs[job_name]
    job.resource_class == "2xlarge"
    not job_name in allowed_large_jobs
    msg := sprintf("Job '%s' is not allowed to use 2xlarge resource class", [job_name])
}

allowed_large_jobs := {"build-production", "integration-tests"}

# Limit parallelism
deny contains msg if {
    job := input.config.jobs[job_name]
    job.parallelism > 10
    msg := sprintf("Job '%s' parallelism cannot exceed 10", [job_name])
}

# Require caching for build jobs
deny contains msg if {
    job := input.config.jobs[job_name]
    startswith(job_name, "build")
    not uses_caching(job)
    msg := sprintf("Build job '%s' must use caching", [job_name])
}

uses_caching(job) if {
    job.steps[_].restore_cache
}
EOT
}

# Compliance policy for sensitive data
resource "circleci_policy" "data_compliance" {
  name        = "Data Compliance"
  description = "Ensure compliance with data protection regulations"
  org_id      = data.circleci_organization.main.id
  enabled     = true

  content = <<EOT
package org

import rego.v1

# Prohibit logging sensitive data
deny contains msg if {
    job := input.config.jobs[job_name]
    step := job.steps[_]
    step.run.command
    contains(step.run.command, "echo $PASSWORD")
    msg := sprintf("Job '%s' must not log sensitive environment variables", [job_name])
}

# Require encrypted environment variables for production
deny contains msg if {
    job := input.config.jobs[job_name]
    contains(job_name, "production")
    env_var := job.environment[var_name]
    var_name in sensitive_vars
    not startswith(env_var, "enc:")
    msg := sprintf("Sensitive variable '%s' must be encrypted in production jobs", [var_name])
}

sensitive_vars := {"DATABASE_PASSWORD", "API_SECRET", "PRIVATE_KEY"}
EOT
}

# List all policies in the organization
data "circleci_policies" "all" {
  org_id = data.circleci_organization.main.id
}

# Output policy information
output "enabled_policies" {
  description = "List of enabled policy names"
  value = [
    for policy in data.circleci_policies.all.policies :
    policy.name if policy.enabled
  ]
}

output "policy_count" {
  description = "Total number of policies"
  value = length(data.circleci_policies.all.policies)
}
EOT
