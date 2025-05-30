# Resource: circleci_environment_variable

Manages environment variables within a CircleCI context.

## Example Usage

```hcl
resource "circleci_context" "shared" {
  name = "shared-context"
  owner = {
    id   = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
    slug = "github"
    type = "organization"
  }
}

resource "circleci_environment_variable" "api_key" {
  context_id = circleci_context.shared.id
  name       = "API_KEY"
  value      = "your-secret-api-key"
}

resource "circleci_environment_variable" "database_url" {
  context_id = circleci_context.shared.id
  name       = "DATABASE_URL"
  value      = var.database_url
}
```

## Argument Reference

The following arguments are supported:

* `context_id` - (Required) The ID of the context to add the environment variable to.
* `name` - (Required) The name of the environment variable.
* `value` - (Required) The value of the environment variable.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier for this environment variable (combination of context_id and name).
* `created_at` - The timestamp when the environment variable was created.

## Import

Environment variables can be imported using the format `context_id:variable_name`:

```bash
terraform import circleci_environment_variable.api_key cb803025-ea3a-4271-af9b-52ee456a5de9:API_KEY
```
