# Data Source: circleci_project

Retrieve information about a CircleCI project.

## Example Usage

```hcl
data "circleci_project" "my_project" {
  slug = "gh/my-org/my-repo"
}

output "project_id" {
  value = data.circleci_project.my_project.id
}

output "default_branch" {
  value = data.circleci_project.my_project.default_branch
}
```

## Argument Reference

The following arguments are supported:

* `slug` - (Required) Project slug in the form `vcs-slug/org-name/repo-name`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the project.
* `name` - The name of the project.
* `organization_name` - The name of the organization that owns the project.
* `organization_slug` - The slug of the organization that owns the project.
* `organization_id` - The ID of the organization that owns the project.
* `vcs_type` - The version control system type (e.g., "github", "bitbucket").
* `vcs_url` - The URL of the project in the version control system.
* `default_branch` - The default branch of the project.
* `setup` - Whether the project has been set up in CircleCI.
* `following` - Whether the current user is following the project.

## Import

Projects can be referenced using their slug:

```hcl
data "circleci_project" "example" {
  slug = "gh/example-org/example-repo"
}
```
