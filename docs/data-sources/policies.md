# circleci_policies

Retrieves information about all policies in a CircleCI organization.

## Example Usage

```hcl
data "circleci_policies" "org_policies" {
  org_id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
}

# Output all enabled policies
output "enabled_policies" {
  value = [
    for policy in data.circleci_policies.org_policies.policies :
    policy.name if policy.enabled
  ]
}

# Use in other resources
resource "circleci_policy" "new_policy" {
  name    = "Additional Security Policy"
  org_id  = data.circleci_policies.org_policies.org_id
  enabled = true
  content = "package org\n# Policy content here"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) The organization ID for which to retrieve policies.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `policies` - A list of policies in the organization. Each policy has the following attributes:
  * `id` - The unique identifier of the policy.
  * `name` - The name of the policy.
  * `description` - The description of the policy.
  * `content` - The policy content in OPA Rego format.
  * `enabled` - Whether the policy is enabled.
  * `created_at` - The date and time the policy was created.
  * `updated_at` - The date and time the policy was last updated.

## Notes

* This data source retrieves all policies in the organization, both enabled and disabled.
* Policy content can be large, so use this data source judiciously in configurations with many policies.
* The policies are returned in the order provided by the CircleCI API.
