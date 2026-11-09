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
