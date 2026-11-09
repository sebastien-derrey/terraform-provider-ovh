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

var _ datasource.DataSourceWithConfigure = (*cloudGatewayModelsDataSource)(nil)

func NewCloudGatewayModelsDataSource() datasource.DataSource {
	return &cloudGatewayModelsDataSource{}
}

type cloudGatewayModelsDataSource struct {
	config *Config
}

// CloudGatewayModelReferenceAPI is publicCloud.reference.gateway.Model.
type CloudGatewayModelReferenceAPI struct {
	Name      string                    `json:"name"`
	Location  cloudReferenceAPILocation `json:"location"`
	Bandwidth int64                     `json:"bandwidth"`
}

type cloudGatewayModelsModel struct {
	ServiceName ovhtypes.TfStringValue `tfsdk:"service_name"`
	Region      ovhtypes.TfStringValue `tfsdk:"region"`
	Name        ovhtypes.TfStringValue `tfsdk:"name"`
	Models      types.List             `tfsdk:"models"`
}

func cloudGatewayModelReferenceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":      ovhtypes.TfStringType{},
		"location":  types.ObjectType{AttrTypes: cloudReferenceLocationAttrTypes()},
		"bandwidth": types.Int64Type,
	}
}

func cloudGatewayModelReferenceToObject(m CloudGatewayModelReferenceAPI) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(cloudGatewayModelReferenceAttrTypes(), map[string]attr.Value{
		"name":      ovhtypes.TfStringValue{StringValue: types.StringValue(m.Name)},
		"location":  buildCloudReferenceLocationObject(m.Location),
		"bandwidth": types.Int64Value(m.Bandwidth),
	})
}

func cloudGatewayModelsToList(models []CloudGatewayModelReferenceAPI) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(models))
	for _, m := range models {
		obj, d := cloudGatewayModelReferenceToObject(m)
		diags.Append(d...)
		items = append(items, obj)
	}
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: cloudGatewayModelReferenceAttrTypes()}), diags
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: cloudGatewayModelReferenceAttrTypes()}, items)
	diags.Append(d...)
	return list, diags
}

func (d *cloudGatewayModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_gateway_models"
}

func (d *cloudGatewayModelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if config := cloudReferenceConfigure(req, resp); config != nil {
		d.config = config
	}
}

func (d *cloudGatewayModelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to list the gateway models sold to a public cloud project and their bandwidth. This is read-only reference data: a model sold in several regions appears once per region.",
		MarkdownDescription: "Use this data source to list the gateway models sold to a public cloud project and their bandwidth. This is read-only reference data: a model sold in several regions appears once per region.",
		Attributes: map[string]schema.Attribute{
			"service_name": cloudReferenceServiceNameAttribute(),
			"region":       cloudReferenceFilterAttribute("Restrict the listing to the models sold in this region."),
			"name":         cloudReferenceFilterAttribute("Restrict the listing to the model with this name (S, M, L, XL, 2XL or 3XL)."),
			"models": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Gateway models, one entry per model and region.",
				MarkdownDescription: "Gateway models, one entry per model and region.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": cloudReferenceComputedString("Model name, to use as the model when creating a gateway"),
						"location": cloudReferenceLocationAttribute(
							"Region where the model is sold to the project, whether that region is already activated on the project, and its availability zones",
						),
						"bandwidth": cloudReferenceComputedInt64("Bandwidth, in Mbit/s"),
					},
				},
			},
		},
	}
}

func (d *cloudGatewayModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudGatewayModelsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{}
	cloudReferenceSetQuery(query, "region", data.Region)
	cloudReferenceSetQuery(query, "name", data.Name)
	endpoint := cloudReferenceEndpoint(data.ServiceName.ValueString(), "reference/gateway/model", query)

	var models []CloudGatewayModelReferenceAPI
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &models); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error calling Get %s", endpoint), err.Error())
		return
	}

	list, diags := cloudGatewayModelsToList(models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Models = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
