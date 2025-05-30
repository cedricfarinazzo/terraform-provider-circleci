# circleci_runner

Manages a CircleCI self-hosted runner.

## Example Usage

```hcl
resource "circleci_runner" "linux_runner" {
  name          = "Linux Build Runner"
  description   = "Ubuntu 22.04 runner for building applications"
  resource_class = "myorg/linux-medium"
}

resource "circleci_runner" "windows_runner" {
  name          = "Windows Build Runner"
  description   = "Windows Server 2022 runner for .NET builds"
  resource_class = "myorg/windows-large"
}

# Output runner information
output "runner_info" {
  value = {
    linux_runner_id = circleci_runner.linux_runner.id
    linux_runner_ip = circleci_runner.linux_runner.ip
    windows_runner_id = circleci_runner.windows_runner.id
    windows_runner_ip = circleci_runner.windows_runner.ip
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the runner.
* `description` - (Optional) A description of the runner.
* `resource_class` - (Required) The resource class for the runner. Changing this forces a new resource to be created.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the runner.
* `platform` - The platform of the runner (linux, windows, darwin).
* `ip` - The IP address of the runner.
* `hostname` - The hostname of the runner.
* `version` - The version of the runner agent.
* `first_connected` - The date and time the runner first connected.
* `last_connected` - The date and time the runner last connected.
* `last_used` - The date and time the runner was last used.
* `state` - The current state of the runner.

## Import

Runners can be imported using their ID:

```
terraform import circleci_runner.example 550e8400-e29b-41d4-a716-446655440000
```

## Notes

* Self-hosted runners provide more control over the build environment and can access private resources.
* Runners must be configured with the appropriate resource class before they can accept jobs.
* The runner agent must be installed and configured separately on the target machine.
* Changing the `resource_class` requires creating a new runner resource.
