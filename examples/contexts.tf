terraform {
  required_providers {
    circleci = {
      source = "your-org/circleci"
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

# Create multiple contexts for different environments
resource "circleci_context" "development" {
  name = "development"
  owner = {
    id   = var.organization_id
    slug = "github"
    type = "organization"
  }
}

resource "circleci_context" "staging" {
  name = "staging"
  owner = {
    id   = var.organization_id
    slug = "github"
    type = "organization"
  }
}

resource "circleci_context" "production" {
  name = "production"
  owner = {
    id   = var.organization_id
    slug = "github"
    type = "organization"
  }
}

# Environment variables for development
resource "circleci_environment_variable" "dev_api_url" {
  context_id = circleci_context.development.id
  name       = "API_URL"
  value      = "https://api-dev.example.com"
}

resource "circleci_environment_variable" "dev_database_url" {
  context_id = circleci_context.development.id
  name       = "DATABASE_URL"
  value      = "postgresql://dev:pass@dev-db.example.com/app"
}

# Environment variables for staging
resource "circleci_environment_variable" "staging_api_url" {
  context_id = circleci_context.staging.id
  name       = "API_URL"
  value      = "https://api-staging.example.com"
}

resource "circleci_environment_variable" "staging_database_url" {
  context_id = circleci_context.staging.id
  name       = "DATABASE_URL"
  value      = "postgresql://staging:pass@staging-db.example.com/app"
}

# Environment variables for production
resource "circleci_environment_variable" "prod_api_url" {
  context_id = circleci_context.production.id
  name       = "API_URL"
  value      = "https://api.example.com"
}

resource "circleci_environment_variable" "prod_database_url" {
  context_id = circleci_context.production.id
  name       = "DATABASE_URL"
  value      = "postgresql://prod:securepass@prod-db.example.com/app"
}

# Shared environment variables across all contexts
locals {
  shared_env_vars = {
    "COMPANY_NAME" = "Example Corp"
    "SUPPORT_EMAIL" = "support@example.com"
    "LOG_LEVEL" = "info"
  }
}

resource "circleci_environment_variable" "dev_shared" {
  for_each = local.shared_env_vars
  
  context_id = circleci_context.development.id
  name       = each.key
  value      = each.value
}

resource "circleci_environment_variable" "staging_shared" {
  for_each = local.shared_env_vars
  
  context_id = circleci_context.staging.id
  name       = each.key
  value      = each.value
}

resource "circleci_environment_variable" "prod_shared" {
  for_each = local.shared_env_vars
  
  context_id = circleci_context.production.id
  name       = each.key
  value      = each.value
}

# Outputs
output "context_ids" {
  description = "Context IDs for different environments"
  value = {
    development = circleci_context.development.id
    staging     = circleci_context.staging.id
    production  = circleci_context.production.id
  }
}
