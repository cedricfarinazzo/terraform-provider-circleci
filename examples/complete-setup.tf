# Complete CircleCI Infrastructure Management Example

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

# Organization data
data "circleci_organization" "main" {
  slug = "github"
  name = "your-org"
}

# Create a shared context for common environment variables
resource "circleci_context" "shared" {
  name = "shared-infrastructure"
  owner = {
    id   = data.circleci_organization.main.id
    slug = data.circleci_organization.main.slug
    type = "organization"
  }
}

# Add environment variables to the context
resource "circleci_environment_variable" "api_key" {
  context_id = circleci_context.shared.id
  name       = "API_KEY"
  value      = var.api_key
}

resource "circleci_environment_variable" "database_url" {
  context_id = circleci_context.shared.id
  name       = "DATABASE_URL"
  value      = var.database_url
}

# Set up a project
resource "circleci_project" "main_app" {
  slug = "gh/your-org/main-app"
}

# Add SSH key for deployments
resource "circleci_checkout_key" "deploy_key" {
  project_slug = circleci_project.main_app.slug
  type         = "deploy-key"
}

# Configure webhook for build notifications
resource "circleci_webhook" "slack_notifications" {
  name   = "Slack Build Notifications"
  url    = var.slack_webhook_url
  events = ["workflow-completed"]
  
  scope = {
    id   = circleci_project.main_app.id
    type = "project"
  }
  
  verify_tls     = true
  signing_secret = var.webhook_secret
}

# Schedule nightly builds
resource "circleci_schedule" "nightly" {
  project_slug = circleci_project.main_app.slug
  name         = "Nightly Security Scan"
  description  = "Run security scans every night"
  
  timetable = {
    hours_of_day = [2]
    days_of_week = ["MON", "TUE", "WED", "THU", "FRI"]
  }
  
  attribution_actor = {
    id = var.user_id
  }
  
  parameters = {
    run_security_scan = true
  }
}

# Security policy
resource "circleci_policy" "security_requirements" {
  name        = "Security Requirements"
  description = "Enforce security scanning and approval workflows"
  org_id      = data.circleci_organization.main.id
  enabled     = true

  content = <<EOT
package org

import rego.v1

# Require security scan job in production workflows
deny contains msg if {
    input.config.workflows[workflow_name].jobs[_].name == "deploy-production"
    not has_security_scan(workflow_name)
    msg := sprintf("Workflow '%s' must include security scanning before production deployment", [workflow_name])
}

has_security_scan(workflow_name) if {
    input.config.workflows[workflow_name].jobs[_].name == "security-scan"
}

# Require manual approval for production deployments
deny contains msg if {
    input.config.workflows[_].jobs[job].name == "deploy-production"
    not input.config.jobs[job].type == "approval"
    msg := "Production deployments must require manual approval"
}
EOT
}

# Monthly usage export
resource "circleci_usage_export" "monthly" {
  org_id = data.circleci_organization.main.id
  start  = "${formatdate("YYYY-MM", timestamp())}-01T00:00:00Z"
  end    = "${formatdate("YYYY-MM", timeadd(timestamp(), "24h"))}-01T00:00:00Z"
}

# Variables
variable "api_key" {
  description = "API key for external services"
  type        = string
  sensitive   = true
}

variable "database_url" {
  description = "Database connection URL"
  type        = string
  sensitive   = true
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for notifications"
  type        = string
}

variable "webhook_secret" {
  description = "Secret for webhook signature verification"
  type        = string
  sensitive   = true
}

variable "user_id" {
  description = "User ID for schedule attribution"
  type        = string
}

# Outputs
output "context_id" {
  description = "ID of the shared context"
  value       = circleci_context.shared.id
}

output "project_slug" {
  description = "Project slug for the main application"
  value       = circleci_project.main_app.slug
}

output "usage_export_url" {
  description = "URL to download the monthly usage report"
  value       = circleci_usage_export.monthly.download_url
  sensitive   = true
}

output "security_policy_id" {
  description = "ID of the security policy"
  value       = circleci_policy.security_requirements.id
}
