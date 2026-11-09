package ovh

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

var _ datasource.DataSourceWithConfigure = (*cloudProductAvailabilityDataSource)(nil)

func NewCloudProductAvailabilityDataSource() datasource.DataSource {
	return &cloudProductAvailabilityDataSource{}
}

type cloudProductAvailabilityDataSource struct {
	config *Config
}

func (d *cloudProductAvailabilityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_product_availability"
}

func (d *cloudProductAvailabilityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if config := cloudReferenceConfigure(req, resp); config != nil {
		d.config = config
	}
}

func (d *cloudProductAvailabilityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to list the products available to a public cloud project, one entry per product model and location. This is read-only reference data computed from the OVHcloud commercial catalog combined with the project's entitlements and quota usage: it does not reflect existing resources.",
		MarkdownDescription: "Use this data source to list the products available to a public cloud project, one entry per product model and location. This is read-only reference data computed from the OVHcloud commercial catalog combined with the project's entitlements and quota usage: it does not reflect existing resources.",
		Attributes: map[string]schema.Attribute{
			"service_name":  cloudReferenceServiceNameAttribute(),
			"type":          cloudReferenceFilterAttribute("Keep the entries of this product type only (for example INSTANCE, VOLUME, LOADBALANCER, GATEWAY)."),
			"region":        cloudReferenceFilterAttribute("Keep the entries of this region only, letter case ignored."),
			"name":          cloudReferenceFilterAttribute("Keep the entries of the models with exactly this name, for example b3-8."),
			"name_contains": cloudReferenceFilterAttribute("Keep the entries of the models whose name contains this value, letter case ignored."),
			"availabilities": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Availability of each product model, one entry per model and location, grouped by product type. A model available in no location has no entry.",
				MarkdownDescription: "Availability of each product model, one entry per model and location, grouped by product type. A model available in no location has no entry.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":   cloudReferenceComputedString("Product type: INSTANCE, VOLUME, LOADBALANCER, GATEWAY, FLOATING_IP, PUBLIC_IP, VOLUME_SNAPSHOT, VOLUME_BACKUP, INSTANCE_SNAPSHOT, SHARE, SHARE_SNAPSHOT, OBJECT_STORAGE or SWIFT."),
						"name":   cloudReferenceComputedString("Model name within its product type, to use when creating it: the flavor name for an instance or a load balancer, the volume type for a volume, the model for a gateway, the share type for file storage, the storage class for object storage. Products sold as a single model (floating IP, public IP, snapshots, backups, Swift) use their product type as model name."),
						"region": cloudReferenceComputedString("Region or location identifier, for example GRA11."),
						"region_activated": schema.BoolAttribute{
							Computed:            true,
							Description:         "Whether the region is already activated on the project. When false, the model is sold in the location but the region must be activated on the project before use. This says nothing about stock or quota.",
							MarkdownDescription: "Whether the region is already activated on the project. When false, the model is sold in the location but the region must be activated on the project before use. This says nothing about stock or quota.",
						},
						"availability_zones": schema.ListAttribute{
							ElementType:         ovhtypes.TfStringType{},
							Computed:            true,
							Description:         "Availability zones within the location, for example eu-west-par-a, eu-west-par-b, eu-west-par-c. Empty for single-AZ regions.",
							MarkdownDescription: "Availability zones within the location, for example `eu-west-par-a`, `eu-west-par-b`, `eu-west-par-c`. Empty for single-AZ regions.",
						},
						"stock_status": schema.MapAttribute{
							ElementType:         ovhtypes.TfStringType{},
							Computed:            true,
							Description:         "Current stock status of the model in the location (AVAILABLE, LOW or OUT_OF_STOCK), keyed by availability-zone name. A 3AZ region has one entry per AZ; single-AZ and localzone regions have a single entry. Filled only for the INSTANCE product type; empty for every other product type.",
							MarkdownDescription: "Current stock status of the model in the location (`AVAILABLE`, `LOW` or `OUT_OF_STOCK`), keyed by availability-zone name. A 3AZ region has one entry per AZ; single-AZ and localzone regions have a single entry. Filled only for the `INSTANCE` product type; empty for every other product type.",
						},
						"quota_status": schema.StringAttribute{
							CustomType:          ovhtypes.TfStringType{},
							Computed:            true,
							Description:         "Whether the project's quota in the location leaves room for one more unit of the model: SUFFICIENT or INSUFFICIENT, read from the last quota usage snapshot, which can lag a few minutes behind. SUFFICIENT as well for a product with no tracked quota in a region activated on the project. UNKNOWN when the region is not activated on the project or no quota usage is known for it.",
							MarkdownDescription: "Whether the project's quota in the location leaves room for one more unit of the model: `SUFFICIENT` or `INSUFFICIENT`, read from the last quota usage snapshot, which can lag a few minutes behind. `SUFFICIENT` as well for a product with no tracked quota in a region activated on the project. `UNKNOWN` when the region is not activated on the project or no quota usage is known for it.",
						},
						"quota_limited_by": schema.ListAttribute{
							ElementType:         ovhtypes.TfStringType{},
							Computed:            true,
							Description:         "Quota resources without enough room (for example CORES, RAM, INSTANCES), when quota_status is INSUFFICIENT. Empty otherwise.",
							MarkdownDescription: "Quota resources without enough room (for example `CORES`, `RAM`, `INSTANCES`), when `quota_status` is `INSUFFICIENT`. Empty otherwise.",
						},
					},
				},
			},
		},
	}
}

func cloudProductAvailabilityEndpoint(data *CloudProductAvailabilityModel) string {
	query := url.Values{}
	cloudReferenceSetQuery(query, "type", data.Type)
	cloudReferenceSetQuery(query, "region", data.Region)
	cloudReferenceSetQuery(query, "name", data.Name)
	cloudReferenceSetQuery(query, "nameContains", data.NameContains)
	return cloudReferenceEndpoint(data.ServiceName.ValueString(), "reference/productAvailability", query)
}

func (d *cloudProductAvailabilityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudProductAvailabilityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := cloudProductAvailabilityEndpoint(&data)

	var responseData []CloudProductAvailabilityAPIResponse
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &responseData); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error calling Get %s", endpoint), err.Error())
		return
	}

	availabilities, diags := cloudProductAvailabilityFlatten(responseData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Availabilities = availabilities

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
