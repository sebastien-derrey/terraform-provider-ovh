---
subcategory : "Cloud Project"
---

# ovh_cloud_region (Data Source)

Use this data source to retrieve information about a single region of a public cloud project, using the OVHcloud API v2.

## Example Usage

```terraform
data "ovh_cloud_region" "region" {
  service_name = "<public cloud project ID>"
  name         = "GRA11"
}
```

## Argument Reference

* `service_name` - (Required) The id of the public cloud project.
* `name`         - (Required) The name of the region (e.g. `GRA11`).

## Attributes Reference

The following attributes are exported:

* `status`             - Region status (`ENABLED`, `DISABLED` or `MAINTENANCE`).
* `continent`          - Continent code of the region.
* `country`            - Country code of the region.
* `datacenter_name`    - Display name of the datacenter hosting the region.
* `availability_zones` - Availability zones available in the region.
* `services`           - Available OpenStack services in the region.
* `type`               - Region type (`LOCAL-ZONE`, `REGION-1-AZ` or `REGION-3-AZ`); null when the catalog reports an unrecognized type.
* `activated`          - Whether the region is activated on the project. When false, the region exists but is not yet activated for the project.
* `object_storage`     - Object storage capabilities of the region; null when the region offers no object storage:
  * `bare_container_only`      - Whether the region only supports bare containers, without the additional bucket features (encryption, versioning, object lock, lifecycle, replication).
  * `is_replication_target`    - Whether a bucket in this region can be used as the destination of a replication rule.
  * `lifecycle_target_classes` - Storage classes that can be used as a transition target in a lifecycle rule in this region.
  * `replication_supported`    - Whether replication rules can be configured on buckets located in this region.
  * `storage_classes`          - Storage classes orderable in this region.
