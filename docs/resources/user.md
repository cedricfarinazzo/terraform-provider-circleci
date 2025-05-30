# circleci_user

Manages a CircleCI organization user. This resource allows you to invite users to your organization and manage their roles.

## Example Usage

```hcl
resource "circleci_user" "developer" {
  email  = "developer@example.com"
  org_id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
  role   = "member"
}

resource "circleci_user" "admin" {
  email  = "admin@example.com"
  org_id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
  role   = "admin"
}
```

## Argument Reference

The following arguments are supported:

* `email` - (Required) The email address of the user to invite.
* `org_id` - (Required) The organization ID where the user will be added. Changing this forces a new resource to be created.
* `role` - (Required) The role to assign to the user. Valid values are `admin`, `member`, and `viewer`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the user.
* `login` - The user's login name.
* `name` - The user's full name.
* `avatar_url` - The user's avatar URL.
* `joined_at` - The date and time the user joined the organization.

## Import

Users can be imported using the organization ID and user ID separated by a colon:

```
terraform import circleci_user.example bb604b45-b6b0-4b81-ad80-796f15eddf87:550e8400-e29b-41d4-a716-446655440000
```

## Notes

* When a user resource is destroyed, the user will be removed from the organization.
* Users must accept the invitation to join the organization before they appear as active members.
* Role changes take effect immediately for existing users.
