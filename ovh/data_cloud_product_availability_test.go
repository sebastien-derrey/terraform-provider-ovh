package ovh

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceCloudProductAvailability_basic(t *testing.T) {
	serviceName := os.Getenv("OVH_CLOUD_PROJECT_SERVICE_TEST")
	region := os.Getenv("OVH_CLOUD_PROJECT_REGION_TEST")

	config := fmt.Sprintf(`
data "ovh_cloud_product_availability" "instances" {
  service_name = "%s"
  type         = "INSTANCE"
  region       = "%s"
}
`, serviceName, region)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCloudInstanceV2(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ovh_cloud_product_availability.instances", "availabilities.#"),
					resource.TestCheckResourceAttr("data.ovh_cloud_product_availability.instances", "availabilities.0.type", "INSTANCE"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_product_availability.instances", "availabilities.0.name"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_product_availability.instances", "availabilities.0.region"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_product_availability.instances", "availabilities.0.region_activated"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_product_availability.instances", "availabilities.0.quota_status"),
				),
			},
		},
	})
}
