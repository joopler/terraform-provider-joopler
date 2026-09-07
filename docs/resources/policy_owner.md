---
page_title: "joopler_policy_owner Resource - joopler"
subcategory: ""
description: |-
  Assign the owner of a policy. Sign-off remains a human act.
---

# joopler_policy_owner (Resource)

Assign the owner of a policy. The owner (or an admin) still signs off on the
policy as a human act - this only assigns accountability.

## Example usage

```terraform
resource "joopler_policy_owner" "aup" {
  policy_key = "acceptable-use"
  owner      = "legal@acme.com"
}
```

## Schema

### Required

- `policy_key` (String) Policy key (e.g. `acceptable-use`).
- `owner` (String) Email of the accountable owner.
