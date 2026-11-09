data "ovh_cloud_gateway_models" "models" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "gateway_bandwidth_by_model" {
  value = { for m in data.ovh_cloud_gateway_models.models.models : m.name => m.bandwidth }
}
