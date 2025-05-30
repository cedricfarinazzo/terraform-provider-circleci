# circleci_runner_token

Manages a CircleCI runner authentication token.

## Example Usage

```hcl
resource "circleci_runner_token" "linux_token" {
  resource_class = "myorg/linux-medium"
  nickname       = "Linux Runner Token"
}

resource "circleci_runner_token" "windows_token" {
  resource_class = "myorg/windows-large"
  nickname       = "Windows Runner Token"
}

# Output tokens for runner configuration
output "runner_tokens" {
  value = {
    linux_token   = circleci_runner_token.linux_token.token
    windows_token = circleci_runner_token.windows_token.token
  }
  sensitive = true
}

# Use token in runner configuration script
resource "local_file" "runner_config" {
  content = templatefile("${path.module}/runner-config.sh.tpl", {
    token = circleci_runner_token.linux_token.token
    resource_class = circleci_runner_token.linux_token.resource_class
  })
  filename = "${path.module}/runner-config.sh"
}
```

## Argument Reference

The following arguments are supported:

* `resource_class` - (Required) The resource class for which this token provides access. Changing this forces a new resource to be created.
* `nickname` - (Required) A human-readable name for the token. Changing this forces a new resource to be created.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the runner token.
* `token` - The authentication token value. This is only available when the token is first created.
* `created_at` - The date and time the token was created.

## Import

Runner tokens can be imported using their ID:

```
terraform import circleci_runner_token.example 550e8400-e29b-41d4-a716-446655440000
```

## Notes

* Runner tokens are immutable once created. Any changes require creating a new token.
* The `token` value is only returned when the token is first created and cannot be retrieved later.
* Tokens should be stored securely and rotated regularly for security.
* Each token is associated with a specific resource class and cannot be used for other resource classes.
* When a token resource is destroyed, the token is immediately revoked and cannot be used by runners.
