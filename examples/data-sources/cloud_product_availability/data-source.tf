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
