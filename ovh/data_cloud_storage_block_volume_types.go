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

var _ datasource.DataSourceWithConfigure = (*cloudStorageBlockVolumeTypesDataSource)(nil)

func NewCloudStorageBlockVolumeTypesDataSource() datasource.DataSource {
	return &cloudStorageBlockVolumeTypesDataSource{}
}

type cloudStorageBlockVolumeTypesDataSource struct {
	config *Config
}

// CloudStorageBlockVolumeTypeReferenceAPI is publicCloud.reference.storage.block.VolumeType.
type CloudStorageBlockVolumeTypeReferenceAPI struct {
	Name        string                        `json:"name"`
	Location    cloudReferenceAPILocation     `json:"location"`
	Iops        cloudReferencePerformanceAPI  `json:"iops"`
	Throughput  *cloudReferencePerformanceAPI `json:"throughput"`
	MaximumSize int64                         `json:"maximumSize"`
}

type cloudStorageBlockVolumeTypesModel struct {
	ServiceName ovhtypes.TfStringValue `tfsdk:"service_name"`
	Region      ovhtypes.TfStringValue `tfsdk:"region"`
	VolumeTypes types.List             `tfsdk:"volume_types"`
}

func cloudStorageBlockVolumeTypeReferenceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":         ovhtypes.TfStringType{},
		"location":     types.ObjectType{AttrTypes: cloudReferenceLocationAttrTypes()},
		"iops":         types.ObjectType{AttrTypes: cloudReferencePerformanceAttrTypes()},
		"throughput":   types.ObjectType{AttrTypes: cloudReferencePerformanceAttrTypes()},
		"maximum_size": types.Int64Type,
	}
}

func cloudStorageBlockVolumeTypeReferenceToObject(v CloudStorageBlockVolumeTypeReferenceAPI) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(cloudStorageBlockVolumeTypeReferenceAttrTypes(), map[string]attr.Value{
		"name":         ovhtypes.TfStringValue{StringValue: types.StringValue(v.Name)},
		"location":     buildCloudReferenceLocationObject(v.Location),
		"iops":         buildCloudReferencePerformanceObject(&v.Iops),
		"throughput":   buildCloudReferencePerformanceObject(v.Throughput),
		"maximum_size": types.Int64Value(v.MaximumSize),
	})
}

func cloudStorageBlockVolumeTypesToList(volumeTypes []CloudStorageBlockVolumeTypeReferenceAPI) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(volumeTypes))
	for _, v := range volumeTypes {
		obj, d := cloudStorageBlockVolumeTypeReferenceToObject(v)
		diags.Append(d...)
		items = append(items, obj)
	}
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: cloudStorageBlockVolumeTypeReferenceAttrTypes()}), diags
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: cloudStorageBlockVolumeTypeReferenceAttrTypes()}, items)
	diags.Append(d...)
	return list, diags
}

func (d *cloudStorageBlockVolumeTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_storage_block_volume_types"
}

func (d *cloudStorageBlockVolumeTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if config := cloudReferenceConfigure(req, resp); config != nil {
		d.config = config
	}
}

func (d *cloudStorageBlockVolumeTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to list the block storage volume types sold to a public cloud project, with their performance and size limits. This is read-only reference data: a volume type sold in several regions appears once per region.",
		MarkdownDescription: "Use this data source to list the block storage volume types sold to a public cloud project, with their performance and size limits. This is read-only reference data: a volume type sold in several regions appears once per region.",
		Attributes: map[string]schema.Attribute{
			"service_name": cloudReferenceServiceNameAttribute(),
			"region":       cloudReferenceFilterAttribute("Restrict the listing to the volume types sold in this region."),
			"volume_types": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Volume types, one entry per volume type and region.",
				MarkdownDescription: "Volume types, one entry per volume type and region.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": cloudReferenceComputedString("Volume type identifier, to use as the volume type when creating a volume"),
						"location": cloudReferenceLocationAttribute(
							"Region where the volume type is sold to the project, whether that region is already activated on the project, and its availability zones",
						),
						"iops":         cloudReferencePerformanceAttribute("Input/output operations per second of a volume of this type in this region"),
						"throughput":   cloudReferencePerformanceAttribute("Throughput of a volume of this type in this region; null when no throughput figure is published for this region"),
						"maximum_size": cloudReferenceComputedInt64("Maximum size of a volume of this type in this region, in GB"),
					},
				},
			},
		},
	}
}

func (d *cloudStorageBlockVolumeTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudStorageBlockVolumeTypesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{}
	cloudReferenceSetQuery(query, "region", data.Region)
	endpoint := cloudReferenceEndpoint(data.ServiceName.ValueString(), "reference/storage/block/volumeType", query)

	var volumeTypes []CloudStorageBlockVolumeTypeReferenceAPI
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &volumeTypes); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error calling Get %s", endpoint), err.Error())
		return
	}

	list, diags := cloudStorageBlockVolumeTypesToList(volumeTypes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.VolumeTypes = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
