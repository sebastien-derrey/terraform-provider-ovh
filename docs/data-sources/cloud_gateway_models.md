---
subcategory: "Gateway"
---

# ovh_cloud_gateway_models (Data Source)

Use this data source to list the gateway models sold to a public cloud project and their bandwidth. This is read-only reference data: a model sold in several regions appears once per region.

## Example Usage

```terraform
data "ovh_cloud_gateway_models" "models" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "gateway_bandwidth_by_model" {
  value = { for m in data.ovh_cloud_gateway_models.models.models : m.name => m.bandwidth }
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) Service name of the resource representing the id of the cloud project.
* `region` - (Optional) Restrict the listing to the models sold in this region.
* `name` - (Optional) Restrict the listing to the model with this name (`S`, `M`, `L`, `XL`, `2XL` or `3XL`).

## Attributes Reference

The following attributes are exported:

* `models` - Gateway models, one entry per model and region:
  * `name` - Model name, to use as the model when creating a gateway.
  * `location` - Region where the model is sold to the project, whether that region is already activated on the project, and its availability zones:
    * `region` - Region name, for example `GRA11`.
    * `region_activated` - Whether the region is already activated on the project. When false, the region must be activated on the project before use.
    * `availability_zones` - Availability zones of the region (for example `eu-west-par-a`, `eu-west-par-b`, `eu-west-par-c`). Empty for a single-AZ region.
  * `bandwidth` - Bandwidth, in Mbit/s.
