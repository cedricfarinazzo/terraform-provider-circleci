# Data Source: circleci_organization

Retrieve information about a CircleCI organization.

## Example Usage

```hcl
data "circleci_organization" "my_org" {
  name = "my-organization"
}

output "org_id" {
  value = data.circleci_organization.my_org.id
}

output "org_vcs_type" {
  value = data.circleci_organization.my_org.vcs_type
}
```

## Example Usage by ID

```hcl
data "circleci_organization" "by_id" {
  id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
}
```

## Argument Reference

One of the following arguments must be specified:

* `name` - (Optional) The name of the organization to retrieve.
* `id` - (Optional) The ID of the organization to retrieve.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the organization.
* `name` - The name of the organization.
* `slug` - The slug of the organization.
* `vcs_type` - The version control system type (e.g., "github", "bitbucket").
* `avatar_url` - The URL of the organization's avatar image.

## Import

Organizations can be referenced using their ID or name:

```hcl
data "circleci_organization" "example" {
  name = "example-org"
}
```
