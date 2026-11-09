---
subcategory: "Cloud Project"
---

# ovh_cloud_product_availability (Data Source)

Use this data source to list the products available to a public cloud project, one entry per product model and location. This is read-only reference data computed from the OVHcloud commercial catalog combined with the project's entitlements and quota usage: it does not reflect existing resources.

## Example Usage

```terraform
# Every filter is optional: omit them all to list every product type, model and location.
data "ovh_cloud_product_availability" "instances" {
  service_name = "<Public cloud project id>"
  type         = "INSTANCE"
  name         = "b3-8"
}

# Regions where a b3-8 instance can be created right now: region activated,
# quota headroom and stock in at least one availability zone.
output "b3_8_ready_regions" {
  value = sort([
    for a in data.ovh_cloud_product_availability.instances.availabilities :
    a.region if a.region_activated && a.quota_status == "SUFFICIENT" && contains(values(a.stock_status), "AVAILABLE")
  ])
}

# Regions where the quota blocks a new b3-8 instance, with the blocking quota resources.
output "b3_8_quota_blocked" {
  value = {
    for a in data.ovh_cloud_product_availability.instances.availabilities :
    a.region => a.quota_limited_by if a.quota_status == "INSUFFICIENT"
  }
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) Service name of the resource representing the id of the cloud project.
* `type` - (Optional) Keep the entries of this product type only. One of `INSTANCE`, `VOLUME`, `LOADBALANCER`, `GATEWAY`, `FLOATING_IP`, `PUBLIC_IP`, `VOLUME_SNAPSHOT`, `VOLUME_BACKUP`, `INSTANCE_SNAPSHOT`, `SHARE`, `SHARE_SNAPSHOT`, `OBJECT_STORAGE`, `SWIFT`.
* `region` - (Optional) Keep the entries of this region only, letter case ignored.
* `name` - (Optional) Keep the entries of the models with exactly this name, for example `b3-8`.
* `name_contains` - (Optional) Keep the entries of the models whose name contains this value, letter case ignored.

## Attributes Reference

The following attributes are exported:

* `availabilities` - Availability of each product model, one entry per model and location, grouped by product type in this order: `INSTANCE`, `VOLUME`, `LOADBALANCER`, `GATEWAY`, `FLOATING_IP`, `PUBLIC_IP`, `VOLUME_SNAPSHOT`, `VOLUME_BACKUP`, `INSTANCE_SNAPSHOT`, `SHARE`, `SHARE_SNAPSHOT`, `OBJECT_STORAGE`, `SWIFT`. A model available in no location has no entry:
  * `type` - Product type of the model.
  * `name` - Model name within its product type, to use when creating it: the flavor name for an instance (`b3-8`, `win-b3-8`) or a load balancer (`SMALL` to `XL`), the volume type for a volume (`CLASSIC`, `HIGH_SPEED`, `HIGH_SPEED_GEN2`), the model for a gateway (`S` to `3XL`), the share type for file storage (`STANDARD_1AZ`), the storage class for object storage (`STANDARD`, `STANDARD_IA`, `HIGH_PERF`, `GLACIER_IR`, `DEEP_ARCHIVE`). Products sold as a single model (floating IP, public IP, snapshots, backups, Swift) use their product type as model name.
  * `region` - Region or location identifier, for example `GRA11`.
  * `region_activated` - Whether the region is already activated on the project. When false, the model is sold in the location but the region must be activated on the project before use. This says nothing about stock or quota.
  * `availability_zones` - Availability zones within the location, for example `eu-west-par-a`, `eu-west-par-b`, `eu-west-par-c`. Empty for single-AZ regions.
  * `stock_status` - Current stock status of the model in the location (`AVAILABLE`, `LOW` or `OUT_OF_STOCK`), keyed by availability-zone name. A 3AZ region has one entry per AZ; single-AZ and localzone regions have a single entry. Filled only for the `INSTANCE` product type; empty for every other product type.
  * `quota_status` - Whether the project's quota in the location leaves room for one more unit of the model: one instance with its vCPUs and RAM, one volume, volume snapshot or volume backup of 1 GB, one load balancer, gateway or floating IP, one file storage share of the share type's minimum size, one file storage snapshot. `SUFFICIENT` or `INSUFFICIENT` is read from the last quota usage snapshot, which can lag a few minutes behind. `SUFFICIENT` as well for a product with no tracked quota (public IP, instance snapshot, object storage, Swift) in a region activated on the project. `UNKNOWN` when the region is not activated on the project or no quota usage is known for it.
  * `quota_limited_by` - Quota resources without enough room, when `quota_status` is `INSUFFICIENT`. Empty otherwise. Values: `CORES`, `RAM`, `INSTANCES`, `VOLUMES`, `VOLUME_GIGABYTES`, `VOLUME_SNAPSHOTS`, `VOLUME_BACKUPS`, `VOLUME_BACKUP_GIGABYTES`, `LOADBALANCERS`, `GATEWAYS`, `FLOATING_IPS`, `SHARES`, `SHARE_GIGABYTES`, `SHARE_SNAPSHOTS`.
