terraform {
  required_providers {
    circleci = {
      source = "cedric-farinazzo/circleci"
      version = "~> 0.1.2"
    }
  }
}

provider "circleci" {
  api_token = var.circleci_token
  # base_url = "https://circleci.com/api/v2" # Optional, defaults to CircleCI cloud
}

variable "circleci_token" {
  description = "CircleCI API token"
  type        = string
  sensitive   = true
}

variable "organization_id" {
  description = "Organization ID"
  type        = string
}

# Create a context
resource "circleci_context" "shared" {
  name = "shared-context"
  owner = {
    id   = var.organization_id
    slug = "github"
    type = "organization"
  }
}

# Add environment variables to the context
resource "circleci_environment_variable" "api_key" {
  context_id = circleci_context.shared.id
  name       = "API_KEY"
  value      = "your-secret-api-key"
}

resource "circleci_environment_variable" "database_url" {
  context_id = circleci_context.shared.id
  name       = "DATABASE_URL"
  value      = "postgresql://user:pass@localhost/db"
}

# Follow a project
resource "circleci_project" "example" {
  slug = "gh/your-org/your-repo"
}

# Add a checkout key to the project
resource "circleci_checkout_key" "deploy_key" {
  project_slug = circleci_project.example.slug
  type         = "deploy-key"
}

# Create a webhook
resource "circleci_webhook" "build_notifications" {
  name   = "Build Notifications"
  url    = "https://your-app.com/webhooks/circleci"
  events = ["workflow-completed"]
  
  scope = {
    id   = circleci_project.example.id
    type = "project"
  }
  
  verify_tls      = true
  signing_secret  = "your-webhook-secret"
}

# Create a schedule
resource "circleci_schedule" "nightly_build" {
  project_slug = circleci_project.example.slug
  name         = "Nightly Build"
  description  = "Run tests every night at 2 AM"
  
  timetable = {
    hours_of_day = [2]
    days_of_week = ["MON", "TUE", "WED", "THU", "FRI"]
  }
  
  attribution_actor = {
    id = "your-user-id"
  }
  
  parameters = {
    run_tests = "true"
    environment = "staging"
  }
}

# Data sources
data "circleci_context" "existing" {
  name = "existing-context"
}

data "circleci_project" "existing" {
  slug = "gh/your-org/existing-repo"
}

data "circleci_insight" "workflow_metrics" {
  project_slug = data.circleci_project.existing.slug
  workflow     = "build-and-test"
  branch       = "main"
}

data "circleci_organization" "current" {
  name = "your-organization"
}

# Outputs
output "context_id" {
  description = "The ID of the created context"
  value       = circleci_context.shared.id
}

output "project_info" {
  description = "Project information"
  value = {
    id           = circleci_project.example.id
    name         = circleci_project.example.name
    organization = circleci_project.example.organization
    vcs_url      = circleci_project.example.vcs_url
  }
}

output "workflow_metrics" {
  description = "Workflow performance metrics"
  value = {
    success_rate     = data.circleci_insight.workflow_metrics.metrics.success_rate
    median_duration  = data.circleci_insight.workflow_metrics.metrics.median_duration
    total_runs      = data.circleci_insight.workflow_metrics.metrics.total_runs
  }
}
