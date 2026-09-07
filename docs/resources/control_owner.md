---
page_title: "joopler_control_owner Resource - joopler"
subcategory: ""
description: |-
  Assign the accountable owner of a control.
---

# joopler_control_owner (Resource)

Assign the accountable owner of a control. A failing control routes to its owner
and appears in their **My work** inbox, so ownership is what makes the work
findable rather than a label.

## Example usage

```terraform
resource "joopler_control_owner" "cloudtrail" {
  control_key = "aws-cloudtrail-enabled"
  owner_email = "security@acme.com"
}
```

## Schema

### Required

- `control_key` (String) Control key (e.g. `aws-cloudtrail-enabled`).
- `owner_email` (String) Email of the accountable owner.
