data "ovh_cloud_loadbalancer_flavors" "flavors" {
  service_name = "<public cloud project ID>"
  region       = "GRA11"
}

output "requests_per_second_by_flavor" {
  value = { for f in data.ovh_cloud_loadbalancer_flavors.flavors.flavors : f.name => f.requests_per_second }
}
