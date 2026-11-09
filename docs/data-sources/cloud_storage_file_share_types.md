---
subcategory: "File Storage"
---

# ovh_cloud_storage_file_share_types (Data Source)

Use this data source to list the file storage types sold to a public cloud project, with their protocols, performance and size limits. This is read-only reference data: a file storage type sold in several regions appears once per region.

## Example Usage

```terraform
data "ovh_cloud_storage_file_share_types" "types" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "share_type_size_limits" {
  value = {
    for s in data.ovh_cloud_storage_file_share_types.types.share_types :
    s.name => { min = s.minimum_size, max = s.maximum_size }
  }
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) Service name of the resource representing the id of the cloud project.
* `region` - (Optional) Restrict the listing to the file storage types sold in this region.

## Attributes Reference

The following attributes are exported:

* `share_types` - File storage types, one entry per type and region:
  * `name` - File storage type identifier, to use as the type when creating a file storage.
  * `protocols` - Protocols a file storage of this type supports.
  * `location` - Region where the file storage type is sold to the project, whether that region is already activated on the project, and its availability zones:
    * `region` - Region name, for example `GRA11`.
    * `region_activated` - Whether the region is already activated on the project. When false, the region must be activated on the project before use.
    * `availability_zones` - Availability zones of the region (for example `eu-west-par-a`, `eu-west-par-b`, `eu-west-par-c`). Empty for a single-AZ region.
  * `iops` - Input/output operations per second of a file storage of this type in this region:
    * `guaranteed` - Whether the baseline level is guaranteed.
    * `level` - Baseline performance; may be expressed per GB of size, see its unit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`); units ending in `/GB` scale with the size.
    * `maximum` - Upper performance limit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`).
  * `throughput` - Throughput of a file storage of this type in this region; null when no throughput figure is published for this region:
    * `guaranteed` - Whether the baseline level is guaranteed.
    * `level` - Baseline performance; may be expressed per GB of size, see its unit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`); units ending in `/GB` scale with the size.
    * `maximum` - Upper performance limit:
      * `value` - Numeric value, expressed in `unit`.
      * `unit` - Unit of the value (`IOPS`, `IOPS/GB`, `MB/S` or `MB/S/GB`).
  * `minimum_size` - Minimum size of a file storage of this type in this region, in GB.
  * `maximum_size` - Maximum size of a file storage of this type in this region, in GB.
