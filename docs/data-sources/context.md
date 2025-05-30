# Data Source: circleci_context

Retrieve information about a CircleCI context, including its name, ID, and creation date.

## Example Usage

```hcl
data "circleci_context" "shared" {
  name = "shared-context"
  owner = {
    slug = "github"
    type = "organization"
  }
}

output "context_id" {
  value = data.circleci_context.shared.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the context to retrieve.
* `owner` - (Required) The owner of the context. Structure is documented below.

The `owner` block supports:

* `slug` - (Required) The slug of the owner (e.g., "github").
* `type` - (Required) The type of owner. Must be "organization".
* `id` - (Optional) The ID of the owner organization.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the context.
* `created_at` - The timestamp when the context was created.

## Import

Contexts can be referenced using their ID:

```hcl
data "circleci_context" "example" {
  name = "example-context"
  owner = {
    slug = "github"
    type = "organization"
  }
}
```
