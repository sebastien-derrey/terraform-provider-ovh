---
subcategory: "Block Storage"
---

# ovh_cloud_storage_block_volume_types (Data Source)

Use this data source to list the block storage volume types sold to a public cloud project, with their performance and size limits. This is read-only reference data: a volume type sold in several regions appears once per region.

## Example Usage

```terraform
data "ovh_cloud_storage_block_volume_types" "types" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "volume_type_max_size" {
  value = { for v in data.ovh_cloud_storage_block_volume_types.types.volume_types : v.name => v.maximum_size }
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) Service name of the resource representing the id of the cloud project.
* `region` - (Optional) Restrict the listing to the volume types sold in this region.

## Attributes Reference

The following attributes are exported:

* `volume_types` - Volume types, one entry per volume type and region:
  * `name` - Volume type identifier, to use as the volume type when creating a volume.
  * `location` - Region where the volume type is sold to the project, whether that region is already activated on the project, and its availability zones:
    * `region` - Region name, for example `GRA11`.
    * `region_activated` - Whether the region is already activated on the project. When false, the region must be activated on the project before use.
    * `availability_zones` - Availability zones of the region (for example `eu-west-par-a`, `eu-west-par-b`, `eu-west-par-c`). Empty for a single-AZ region.
  * `iops` - Input/output operations per second of a volume of this type in this region:
    * `guaranteed` - Whether the baseline level is guaranteed.
    * `level` - Baseline performance; may be expressed per GB of size, see its unit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`); units ending in `/GB` scale with the size.
    * `maximum` - Upper performance limit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`).
  * `throughput` - Throughput of a volume of this type in this region; null when no throughput figure is published for this region:
    * `guaranteed` - Whether the baseline level is guaranteed.
    * `level` - Baseline performance; may be expressed per GB of size, see its unit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`); units ending in `/GB` scale with the size.
    * `maximum` - Upper performance limit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`).
  * `maximum_size` - Maximum size of a volume of this type in this region, in GB.
