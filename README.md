# Terraform Provider for CircleCI

A comprehensive Terraform provider for managing CircleCI resources using the CircleCI API v2.

## Features

This provider supports managing the following CircleCI resources:

### Resources
- **Contexts** - Create and manage contexts for sharing environment variables
- **Environment Variables** - Manage environment variables within contexts
- **Projects** - Follow/unfollow projects and manage project settings
- **Checkout Keys** - Manage SSH keys for repository access
- **Webhooks** - Configure webhooks for build notifications
- **Schedules** - Create and manage scheduled pipeline runs
- **OIDC Tokens** - Manage OpenID Connect authentication tokens
- **Policies** - Manage organization policies for compliance and governance
- **Usage Exports** - Export organization usage data for analysis
- **Runners** - Manage self-hosted runners for custom execution environments
- **Runner Tokens** - Manage authentication tokens for self-hosted runners

### Data Sources
- **Context** - Get information about existing contexts
- **Project** - Get information about existing projects
- **Insights** - Retrieve workflow metrics and performance data
- **Organization** - Get information about organizations
- **Policies** - List all policies in an organization

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)

## Installation

### Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    circleci = {
      source  = "your-org/circleci"
      version = "~> 1.0"
    }
  }
}
```

### Manual Installation

1. Download the latest release from the [releases page](https://github.com/cedricfarinazzo/terraform-provider-circleci/releases)
2. Extract the binary to your Terraform plugins directory
3. Run `terraform init`

## Authentication

The provider requires a CircleCI API token for authentication. You can obtain a token from your [CircleCI Personal API Tokens](https://app.circleci.com/settings/user/tokens) page.

### Environment Variable (Recommended)

```bash
export CIRCLECI_TOKEN="your-api-token-here"
```

### Provider Configuration

```hcl
provider "circleci" {
  api_token = "your-api-token-here"
  # base_url = "https://circleci.com/api/v2" # Optional, defaults to CircleCI cloud
}
```

## Usage Examples

### Basic Context Management

```hcl
resource "circleci_context" "shared" {
  name = "shared-context"
  owner = {
    id   = "your-org-id"
    slug = "github"
    type = "organization"
  }
}

resource "circleci_environment_variable" "api_key" {
  context_id = circleci_context.shared.id
  name       = "API_KEY"
  value      = "your-secret-api-key"
}
```

### Project Management

```hcl
resource "circleci_project" "example" {
  slug = "gh/your-org/your-repo"
}

resource "circleci_checkout_key" "deploy_key" {
  project_slug = circleci_project.example.slug
  type         = "deploy-key"
}
```

### Webhook Configuration

```hcl
resource "circleci_webhook" "notifications" {
  name   = "Build Notifications"
  url    = "https://your-app.com/webhooks/circleci"
  events = ["workflow-completed"]
  
  scope = {
    id   = circleci_project.example.id
    type = "project"
  }
  
  verify_tls     = true
  signing_secret = "your-webhook-secret"
}
```

### Scheduled Builds

```hcl
resource "circleci_schedule" "nightly" {
  project_slug = "gh/your-org/your-repo"
  name         = "Nightly Build"
  description  = "Run tests every night"
  
  timetable = {
    hours_of_day = [2]
    days_of_week = ["MON", "TUE", "WED", "THU", "FRI"]
  }
  
  attribution_actor = {
    id = "your-user-id"
  }
}
```

### Data Sources

```hcl
data "circleci_insight" "metrics" {
  project_slug = "gh/your-org/your-repo"
  workflow     = "build-and-test"
  branch       = "main"
}

output "success_rate" {
  value = data.circleci_insight.metrics.metrics.success_rate
}

data "circleci_policies" "org_policies" {
  org_id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
}
```

### Advanced Features

```hcl
# Policy Management
resource "circleci_policy" "security_policy" {
  name        = "Security Requirements"
  description = "Enforce security scanning in all workflows"
  org_id      = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
  enabled     = true
  content     = file("${path.module}/policies/security.rego")
}

# Usage Exports
resource "circleci_usage_export" "monthly_report" {
  org_id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
  start  = "2024-01-01T00:00:00Z"
  end    = "2024-01-31T23:59:59Z"
}

# Self-hosted Runner Management
resource "circleci_runner" "build_runner" {
  resource_class = "my-org/my-runner-class"
  name           = "build-runner-01"
  description    = "Runner for build jobs"
}

resource "circleci_runner_token" "runner_auth" {
  resource_class = "my-org/my-runner-class"
  nickname       = "build-runner-token"
}
```

## Development

### Building the Provider

```bash
git clone https://github.com/cedricfarinazzo/terraform-provider-circleci
cd terraform-provider-circleci
go build
```

### Running Tests

```bash
go test ./...
```

### Running Acceptance Tests

```bash
export CIRCLECI_TOKEN="your-test-token"
export TF_ACC=1
go test ./... -v
```

### Local Development

1. Build the provider:
   ```bash
   go build -o terraform-provider-circleci
   ```

2. Create a `.terraformrc` file in your home directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "your-org/circleci" = "/path/to/terraform-provider-circleci"
     }
     direct {}
   }
   ```

3. Use the provider in your Terraform configuration:
   ```hcl
   terraform {
     required_providers {
       circleci = {
         source = "your-org/circleci"
       }
     }
   }
   ```

## Documentation

Full documentation for all resources and data sources is available in the [docs](./docs) directory or on the [Terraform Registry](https://registry.terraform.io/providers/your-org/circleci/latest/docs).

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Reporting Issues

If you encounter any issues, please report them on the [GitHub issues page](https://github.com/cedricfarinazzo/terraform-provider-circleci/issues).

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/your-org/circleci/latest/docs)
- **Issues**: [GitHub Issues](https://github.com/cedricfarinazzo/terraform-provider-circleci/issues)
- **CircleCI API Documentation**: [CircleCI API v2](https://circleci.com/docs/api/v2/index.html)

## Acknowledgments

- [CircleCI](https://circleci.com/) for providing the API
- [Terraform](https://terraform.io/) for the provider framework
- [HashiCorp](https://hashicorp.com/) for the Terraform Plugin Framework