package ovh

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceCloudLoadbalancerFlavors_basic(t *testing.T) {
	serviceName := os.Getenv("OVH_CLOUD_PROJECT_SERVICE_TEST")
	region := os.Getenv("OVH_CLOUD_PROJECT_REGION_TEST")

	config := fmt.Sprintf(`
data "ovh_cloud_loadbalancer_flavors" "all" {
  service_name = "%s"
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
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.#"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.name"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.bandwidth"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.connections_per_second"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.requests_per_second"),
					resource.TestCheckResourceAttr("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.location.region", region),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.location.region_activated"),
					resource.TestCheckResourceAttrSet("data.ovh_cloud_loadbalancer_flavors.all", "flavors.0.location.availability_zones.#"),
				),
			},
		},
	})
}
