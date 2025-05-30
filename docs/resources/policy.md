# circleci_policy

Manages a CircleCI organizational policy. Policies are written in OPA Rego and allow organizations to enforce compliance rules on their CI/CD workflows.

## Example Usage

```hcl
resource "circleci_policy" "security_policy" {
  name        = "Security Requirements"
  description = "Enforce security scanning in all workflows"
  org_id      = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
  enabled     = true

  content = <<EOT
package org

import rego.v1

# Deny workflows that don't include security scanning
deny contains msg if {
    input.config.workflows[_].jobs[_].name == "build"
    not has_security_scan
    msg := "All build workflows must include security scanning"
}

has_security_scan if {
    input.config.workflows[_].jobs[_].name == "security-scan"
}
EOT
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the policy.
* `description` - (Optional) A description of what the policy enforces.
* `content` - (Required) The policy content written in OPA Rego format.
* `org_id` - (Required) The organization ID where the policy will be applied. Changing this forces a new resource to be created.
* `enabled` - (Required) Whether the policy is enabled and will be enforced.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the policy.
* `created_at` - The date and time the policy was created.
* `updated_at` - The date and time the policy was last updated.

## Import

Policies can be imported using the organization ID and policy ID separated by a colon:

```
terraform import circleci_policy.example bb604b45-b6b0-4b81-ad80-796f15eddf87:550e8400-e29b-41d4-a716-446655440000
```
