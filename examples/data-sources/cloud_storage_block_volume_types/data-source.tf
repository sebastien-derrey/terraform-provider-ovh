data "ovh_cloud_storage_block_volume_types" "types" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "volume_type_max_size" {
  value = { for v in data.ovh_cloud_storage_block_volume_types.types.volume_types : v.name => v.maximum_size }
}
