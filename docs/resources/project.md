# Resource: circleci_project

Manages a CircleCI project by following or unfollowing it.

## Example Usage

```hcl
resource "circleci_project" "my_repo" {
  slug = "gh/my-org/my-repo"
}
```

## Example Usage with Advanced Settings

```hcl
resource "circleci_project" "advanced_repo" {
  slug = "gh/my-org/advanced-repo"
  
  # Additional configuration can be added here
  # when supported by the CircleCI API
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

## Import

Projects can be imported using their slug:

```bash
terraform import circleci_project.my_repo gh/my-org/my-repo
```
