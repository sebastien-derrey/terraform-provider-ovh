---
subcategory: "Load Balancer"
---

# ovh_cloud_loadbalancer_flavors (Data Source)

Use this data source to list the load balancer flavors sold to a public cloud project and their performance. This is read-only reference data: a flavor sold in several regions appears once per region.

## Example Usage

```terraform
data "ovh_cloud_loadbalancer_flavors" "flavors" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "requests_per_second_by_flavor" {
  value = { for f in data.ovh_cloud_loadbalancer_flavors.flavors.flavors : f.name => f.requests_per_second }
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) Service name of the resource representing the id of the cloud project.
* `region` - (Optional) Restrict the listing to the flavors sold in this region.

## Attributes Reference

The following attributes are exported:

* `flavors` - Load balancer flavors, one entry per flavor and region:
  * `name` - Flavor name, to use as the flavor name when creating a load balancer, for example `SMALL`.
  * `location` - Region where the flavor is sold to the project, whether that region is already activated on the project, and its availability zones:
    * `region` - Region name, for example `GRA11`.
    * `region_activated` - Whether the region is already activated on the project. When false, the region must be activated on the project before use.
    * `availability_zones` - Availability zones of the region (for example `eu-west-par-a`, `eu-west-par-b`, `eu-west-par-c`). Empty for a single-AZ region.
  * `bandwidth` - Bandwidth, in Mbit/s.
  * `connections_per_second` - Number of new connections per second the load balancer handles.
  * `requests_per_second` - Number of HTTP requests per second the load balancer handles.
