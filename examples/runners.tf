terraform {
  required_providers {
    circleci = {
      source  = "cedric-farinazzo/circleci"
      version = "~> 0.1.2"
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

variable "organization_slug" {
  description = "Organization slug (e.g., 'my-org')"
  type        = string
}

# Create a self-hosted runner
resource "circleci_runner" "build_runner" {
  resource_class = "${var.organization_slug}/docker-runner"
  name           = "production-build-runner"
  description    = "Self-hosted runner for production builds"
}

# Create authentication token for the runner
resource "circleci_runner_token" "build_runner_token" {
  resource_class = circleci_runner.build_runner.resource_class
  nickname       = "prod-runner-token"
}

# Create another runner for test workloads
resource "circleci_runner" "test_runner" {
  resource_class = "${var.organization_slug}/test-runner"
  name           = "test-environment-runner"
  description    = "Runner dedicated to test execution"
}

resource "circleci_runner_token" "test_runner_token" {
  resource_class = circleci_runner.test_runner.resource_class
  nickname       = "test-runner-token"
}

# Outputs for runner configuration
output "build_runner_id" {
  description = "ID of the build runner"
  value       = circleci_runner.build_runner.id
}

output "build_runner_token" {
  description = "Authentication token for the build runner"
  value       = circleci_runner_token.build_runner_token.token
  sensitive   = true
}

output "test_runner_id" {
  description = "ID of the test runner"
  value       = circleci_runner.test_runner.id
}

output "test_runner_token" {
  description = "Authentication token for the test runner"
  value       = circleci_runner_token.test_runner_token.token
  sensitive   = true
}

# Example of runner setup instructions
output "runner_setup_instructions" {
  description = "Instructions for setting up the runners"
  value = <<-EOT
    To set up your self-hosted runners:
    
    1. Install the CircleCI runner on your infrastructure
    2. Use the following configuration:
    
    Build Runner:
    - Resource Class: ${circleci_runner.build_runner.resource_class}
    - Token: ${circleci_runner_token.build_runner_token.token}
    
    Test Runner:
    - Resource Class: ${circleci_runner.test_runner.resource_class}
    - Token: ${circleci_runner_token.test_runner_token.token}
    
    3. Start the runner agent with your token
  EOT
}
