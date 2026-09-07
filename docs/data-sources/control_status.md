---
page_title: "joopler_control_status Data Source - joopler"
subcategory: ""
description: |-
  Live posture: each control's current status plus rollup counts.
---

# joopler_control_status (Data Source)

Live posture: each control's current status plus rollup counts. Useful for
gating a pipeline on readiness, or publishing the number somewhere your team
already looks.

## Example usage

```terraform
data "joopler_control_status" "this" {}

output "readiness" {
  value = "${data.joopler_control_status.this.passing}/${data.joopler_control_status.this.total} controls passing"
}
```

## Schema

### Read-Only

- `statuses` (Map of String) Control key to status (`PASS` / `FAIL` / `ERROR` / `PENDING`).
- `passing` (Number) Count of passing controls.
- `failing` (Number) Count of failing controls.
- `total` (Number) Total controls.
