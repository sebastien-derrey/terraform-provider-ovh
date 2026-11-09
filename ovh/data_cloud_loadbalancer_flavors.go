package ovh

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

var _ datasource.DataSourceWithConfigure = (*cloudLoadbalancerFlavorsDataSource)(nil)

func NewCloudLoadbalancerFlavorsDataSource() datasource.DataSource {
	return &cloudLoadbalancerFlavorsDataSource{}
}

type cloudLoadbalancerFlavorsDataSource struct {
	config *Config
}

// CloudLoadbalancerFlavorReferenceAPI is publicCloud.reference.loadbalancer.Flavor.
type CloudLoadbalancerFlavorReferenceAPI struct {
	Name                 string                    `json:"name"`
	Location             cloudReferenceAPILocation `json:"location"`
	Bandwidth            int64                     `json:"bandwidth"`
	ConnectionsPerSecond int64                     `json:"connectionsPerSecond"`
	RequestsPerSecond    int64                     `json:"requestsPerSecond"`
}

type cloudLoadbalancerFlavorsModel struct {
	ServiceName ovhtypes.TfStringValue `tfsdk:"service_name"`
	Region      ovhtypes.TfStringValue `tfsdk:"region"`
	Flavors     types.List             `tfsdk:"flavors"`
}

func cloudLoadbalancerFlavorReferenceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                   ovhtypes.TfStringType{},
		"location":               types.ObjectType{AttrTypes: cloudReferenceLocationAttrTypes()},
		"bandwidth":              types.Int64Type,
		"connections_per_second": types.Int64Type,
		"requests_per_second":    types.Int64Type,
	}
}

func cloudLoadbalancerFlavorReferenceToObject(f CloudLoadbalancerFlavorReferenceAPI) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(cloudLoadbalancerFlavorReferenceAttrTypes(), map[string]attr.Value{
		"name":                   ovhtypes.TfStringValue{StringValue: types.StringValue(f.Name)},
		"location":               buildCloudReferenceLocationObject(f.Location),
		"bandwidth":              types.Int64Value(f.Bandwidth),
		"connections_per_second": types.Int64Value(f.ConnectionsPerSecond),
		"requests_per_second":    types.Int64Value(f.RequestsPerSecond),
	})
}

func cloudLoadbalancerFlavorsToList(flavors []CloudLoadbalancerFlavorReferenceAPI) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(flavors))
	for _, f := range flavors {
		obj, d := cloudLoadbalancerFlavorReferenceToObject(f)
		diags.Append(d...)
		items = append(items, obj)
	}
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: cloudLoadbalancerFlavorReferenceAttrTypes()}), diags
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: cloudLoadbalancerFlavorReferenceAttrTypes()}, items)
	diags.Append(d...)
	return list, diags
}

func (d *cloudLoadbalancerFlavorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_loadbalancer_flavors"
}

func (d *cloudLoadbalancerFlavorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if config := cloudReferenceConfigure(req, resp); config != nil {
		d.config = config
	}
}

func (d *cloudLoadbalancerFlavorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to list the load balancer flavors sold to a public cloud project and their performance. This is read-only reference data: a flavor sold in several regions appears once per region.",
		MarkdownDescription: "Use this data source to list the load balancer flavors sold to a public cloud project and their performance. This is read-only reference data: a flavor sold in several regions appears once per region.",
		Attributes: map[string]schema.Attribute{
			"service_name": cloudReferenceServiceNameAttribute(),
			"region":       cloudReferenceFilterAttribute("Restrict the listing to the flavors sold in this region."),
			"flavors": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Load balancer flavors, one entry per flavor and region.",
				MarkdownDescription: "Load balancer flavors, one entry per flavor and region.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": cloudReferenceComputedString("Flavor name, to use as the flavor name when creating a load balancer, for example SMALL"),
						"location": cloudReferenceLocationAttribute(
							"Region where the flavor is sold to the project, whether that region is already activated on the project, and its availability zones",
						),
						"bandwidth":              cloudReferenceComputedInt64("Bandwidth, in Mbit/s"),
						"connections_per_second": cloudReferenceComputedInt64("Number of new connections per second the load balancer handles"),
						"requests_per_second":    cloudReferenceComputedInt64("Number of HTTP requests per second the load balancer handles"),
					},
				},
			},
		},
	}
}

func (d *cloudLoadbalancerFlavorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudLoadbalancerFlavorsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{}
	cloudReferenceSetQuery(query, "region", data.Region)
	endpoint := cloudReferenceEndpoint(data.ServiceName.ValueString(), "reference/loadbalancer/flavor", query)

	var flavors []CloudLoadbalancerFlavorReferenceAPI
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &flavors); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error calling Get %s", endpoint), err.Error())
		return
	}

	list, diags := cloudLoadbalancerFlavorsToList(flavors)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Flavors = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
