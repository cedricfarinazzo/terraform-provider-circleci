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

variable "organization_id" {
  description = "Organization ID"
  type        = string
}

variable "organization_slug" {
  description = "Organization slug"
  type        = string
}

variable "project_slug" {
  description = "Project slug (e.g., 'gh/my-org/my-repo')"
  type        = string
}

# =============================================================================
# RUNNER INFRASTRUCTURE
# =============================================================================

# Create resource classes for different types of workloads
resource "circleci_runner" "build_runner" {
  resource_class = "${var.organization_slug}/build-runner"
  name           = "primary-build-runner"
  description    = "Main runner for building applications"
}

resource "circleci_runner" "test_runner" {
  resource_class = "${var.organization_slug}/test-runner"
  name           = "test-execution-runner"
  description    = "Dedicated runner for test execution"
}

resource "circleci_runner" "deploy_runner" {
  resource_class = "${var.organization_slug}/deploy-runner"
  name           = "production-deployment-runner"
  description    = "Secure runner for production deployments"
}

# Create authentication tokens for each runner
resource "circleci_runner_token" "build_token" {
  resource_class = circleci_runner.build_runner.resource_class
  nickname       = "build-runner-token"
}

resource "circleci_runner_token" "test_token" {
  resource_class = circleci_runner.test_runner.resource_class
  nickname       = "test-runner-token"
}

resource "circleci_runner_token" "deploy_token" {
  resource_class = circleci_runner.deploy_runner.resource_class
  nickname       = "deploy-runner-token"
}

# =============================================================================
# CONTEXT AND SECRETS MANAGEMENT
# =============================================================================

# Create contexts for different environments
resource "circleci_context" "shared" {
  name = "shared-services"
  owner = {
    id   = var.organization_id
    slug = "github"
    type = "organization"
  }
}

resource "circleci_context" "production" {
  name = "production-env"
  owner = {
    id   = var.organization_id
    slug = "github"
    type = "organization"
  }
}

# Add runner tokens to contexts for secure access
resource "circleci_environment_variable" "build_runner_token" {
  context_id = circleci_context.shared.id
  name       = "BUILD_RUNNER_TOKEN"
  value      = circleci_runner_token.build_token.token
}

resource "circleci_environment_variable" "deploy_runner_token" {
  context_id = circleci_context.production.id
  name       = "DEPLOY_RUNNER_TOKEN"
  value      = circleci_runner_token.deploy_token.token
}

# =============================================================================
# PROJECT CONFIGURATION
# =============================================================================

# Follow the project
resource "circleci_project" "main" {
  slug = var.project_slug
}

# Create checkout key for repository access
resource "circleci_checkout_key" "deploy_key" {
  project_slug = circleci_project.main.slug
  type         = "deploy-key"
}

# Configure webhooks for notifications
resource "circleci_webhook" "deployment_notifications" {
  name   = "Deployment Notifications"
  url    = "https://hooks.slack.com/your-webhook-url"
  events = ["workflow-completed"]
  
  scope = {
    id   = circleci_project.main.id
    type = "project"
  }
  
  verify_tls     = true
  signing_secret = "your-webhook-secret"
}

# =============================================================================
# SCHEDULED JOBS
# =============================================================================

# Schedule nightly builds using custom runners
resource "circleci_schedule" "nightly_build" {
  project_slug = var.project_slug
  name         = "Nightly Build and Test"
  description  = "Run comprehensive tests on custom runners every night"
  
  timetable = {
    hours_of_day = [2]  # 2 AM
    days_of_week = ["MON", "TUE", "WED", "THU", "FRI"]
  }
  
  attribution_actor = {
    id = "automation-user-id"
  }
  
  parameters = {
    use_custom_runners = true
    runner_class      = circleci_runner.test_runner.resource_class
  }
}

# =============================================================================
# DATA ANALYSIS AND MONITORING
# =============================================================================

# Get insights for monitoring
data "circleci_insight" "project_metrics" {
  project_slug = var.project_slug
  workflow     = "build-test-deploy"
  branch       = "main"
}

# =============================================================================
# OUTPUTS FOR MONITORING AND AUTOMATION
# =============================================================================

output "runner_configuration" {
  description = "Configuration details for all runners"
  value = {
    build_runner = {
      id             = circleci_runner.build_runner.id
      resource_class = circleci_runner.build_runner.resource_class
      token          = circleci_runner_token.build_token.token
    }
    test_runner = {
      id             = circleci_runner.test_runner.id
      resource_class = circleci_runner.test_runner.resource_class
      token          = circleci_runner_token.test_token.token
    }
    deploy_runner = {
      id             = circleci_runner.deploy_runner.id
      resource_class = circleci_runner.deploy_runner.resource_class
      token          = circleci_runner_token.deploy_token.token
    }
  }
  sensitive = true
}

output "project_metrics" {
  description = "Project performance metrics"
  value = {
    success_rate     = data.circleci_insight.project_metrics.metrics.success_rate
    mean_duration    = data.circleci_insight.project_metrics.metrics.mean_duration_sec
    total_runs       = data.circleci_insight.project_metrics.metrics.total_runs
    throughput       = data.circleci_insight.project_metrics.metrics.throughput
  }
}

# =============================================================================
# RUNNER SETUP INSTRUCTIONS
# =============================================================================

output "runner_setup_script" {
  description = "Script to set up self-hosted runners"
  value = <<-EOT
#!/bin/bash
# CircleCI Self-Hosted Runner Setup Script

echo "Setting up CircleCI Runners..."

# Build Runner Setup
echo "Configuring Build Runner..."
export BUILD_RUNNER_TOKEN="${circleci_runner_token.build_token.token}"
export BUILD_RESOURCE_CLASS="${circleci_runner.build_runner.resource_class}"

# Test Runner Setup  
echo "Configuring Test Runner..."
export TEST_RUNNER_TOKEN="${circleci_runner_token.test_token.token}"
export TEST_RESOURCE_CLASS="${circleci_runner.test_runner.resource_class}"

# Deploy Runner Setup
echo "Configuring Deploy Runner..."
export DEPLOY_RUNNER_TOKEN="${circleci_runner_token.deploy_token.token}"
export DEPLOY_RESOURCE_CLASS="${circleci_runner.deploy_runner.resource_class}"

echo "Runner tokens configured. Start your runner agents with the respective tokens."
echo "Refer to CircleCI documentation for runner installation: https://circleci.com/docs/runner-overview/"
EOT
}
