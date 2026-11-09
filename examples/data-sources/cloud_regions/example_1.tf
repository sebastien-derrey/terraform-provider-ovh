data "ovh_cloud_regions" "regions" {
  service_name = "<public cloud project ID>"
}

output "regions" {
  value = data.ovh_cloud_regions.regions.regions
}

# Regions where buckets can be the destination of a replication rule.
output "replication_target_regions" {
  value = sort([
    for region in data.ovh_cloud_regions.regions.regions :
    region.name if try(region.object_storage.is_replication_target, false)
  ])
}
