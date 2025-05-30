# circleci_usage_export

Manages a CircleCI usage export. This resource allows you to export usage data for your organization over a specified time range.

## Example Usage

```hcl
resource "circleci_usage_export" "monthly_report" {
  org_id = "bb604b45-b6b0-4b81-ad80-796f15eddf87"
  start  = "2024-01-01T00:00:00Z"
  end    = "2024-01-31T23:59:59Z"
}

# Use the export download URL in other resources
output "usage_report_url" {
  value = circleci_usage_export.monthly_report.download_url
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) The organization ID for which to export usage data. Changing this forces a new resource to be created.
* `start` - (Required) The start date for the usage export in ISO 8601 format. Changing this forces a new resource to be created.
* `end` - (Required) The end date for the usage export in ISO 8601 format. Changing this forces a new resource to be created.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the usage export.
* `status` - The status of the export. Possible values are `pending`, `processing`, `completed`, and `failed`.
* `download_url` - The download URL for the completed export (available when status is `completed`).
* `created_at` - The date and time the export was created.
* `expires_at` - The date and time the export download will expire.

## Import

Usage exports can be imported using the organization ID and export ID separated by a colon:

```
terraform import circleci_usage_export.example bb604b45-b6b0-4b81-ad80-796f15eddf87:550e8400-e29b-41d4-a716-446655440000
```

## Notes

* Usage exports are immutable once created. Changes to `start` or `end` dates will force creation of a new export.
* Export processing may take several minutes depending on the data volume.
* Download URLs are temporary and will expire after a certain period.
* The export format is CSV and includes detailed usage metrics for the specified time range.
