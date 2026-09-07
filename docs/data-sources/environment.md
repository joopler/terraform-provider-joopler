---
page_title: "joopler_environment Data Source - joopler"
subcategory: ""
description: |-
  The tenant's connected environment.
---

# joopler_environment (Data Source)

The tenant's connected environment - the observed reality Joopler grounds
policies and the assistant in, rather than a description someone typed once.

## Example usage

```terraform
data "joopler_environment" "this" {}

output "identity_providers" {
  value = data.joopler_environment.this.identity_providers
}
```

## Schema

### Read-Only

- `connected_systems` (List of String) Labels of connected systems.
- `identity_providers` (List of String) Connected identity providers.
- `endpoint_management` (List of String) Connected device-management tools.
- `subprocessor_count` (Number) Number of subprocessors on file.
- `workforce_count` (Number) People on the roster.
