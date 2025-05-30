# Resource: circleci_context

Manages a CircleCI context. Contexts provide a mechanism for securing and sharing environment variables across projects.

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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the context.
* `owner` - (Required) The owner of the context. Structure is documented below.

The `owner` block supports:

* `id` - (Required) The ID of the owner organization.
* `slug` - (Required) The slug of the owner (e.g., "github").
* `type` - (Required) The type of owner. Must be "organization".

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the context.
* `created_at` - The timestamp when the context was created.

## Import

Contexts can be imported using their ID:

```bash
terraform import circleci_context.shared cb803025-ea3a-4271-af9b-52ee456a5de9
```
