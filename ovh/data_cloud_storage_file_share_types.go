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

var _ datasource.DataSourceWithConfigure = (*cloudStorageFileShareTypesDataSource)(nil)

func NewCloudStorageFileShareTypesDataSource() datasource.DataSource {
	return &cloudStorageFileShareTypesDataSource{}
}

type cloudStorageFileShareTypesDataSource struct {
	config *Config
}

// CloudStorageFileShareTypeReferenceAPI is publicCloud.reference.storage.file.ShareType.
type CloudStorageFileShareTypeReferenceAPI struct {
	Name        string                        `json:"name"`
	Protocols   []string                      `json:"protocols"`
	Location    cloudReferenceAPILocation     `json:"location"`
	Iops        cloudReferencePerformanceAPI  `json:"iops"`
	Throughput  *cloudReferencePerformanceAPI `json:"throughput"`
	MinimumSize int64                         `json:"minimumSize"`
	MaximumSize int64                         `json:"maximumSize"`
}

type cloudStorageFileShareTypesModel struct {
	ServiceName ovhtypes.TfStringValue `tfsdk:"service_name"`
	Region      ovhtypes.TfStringValue `tfsdk:"region"`
	ShareTypes  types.List             `tfsdk:"share_types"`
}

func cloudStorageFileShareTypeReferenceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":         ovhtypes.TfStringType{},
		"protocols":    types.ListType{ElemType: ovhtypes.TfStringType{}},
		"location":     types.ObjectType{AttrTypes: cloudReferenceLocationAttrTypes()},
		"iops":         types.ObjectType{AttrTypes: cloudReferencePerformanceAttrTypes()},
		"throughput":   types.ObjectType{AttrTypes: cloudReferencePerformanceAttrTypes()},
		"minimum_size": types.Int64Type,
		"maximum_size": types.Int64Type,
	}
}

func cloudStorageFileShareTypeReferenceToObject(s CloudStorageFileShareTypeReferenceAPI) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(cloudStorageFileShareTypeReferenceAttrTypes(), map[string]attr.Value{
		"name":         ovhtypes.TfStringValue{StringValue: types.StringValue(s.Name)},
		"protocols":    buildStringListFromAPI(s.Protocols),
		"location":     buildCloudReferenceLocationObject(s.Location),
		"iops":         buildCloudReferencePerformanceObject(&s.Iops),
		"throughput":   buildCloudReferencePerformanceObject(s.Throughput),
		"minimum_size": types.Int64Value(s.MinimumSize),
		"maximum_size": types.Int64Value(s.MaximumSize),
	})
}

func cloudStorageFileShareTypesToList(shareTypes []CloudStorageFileShareTypeReferenceAPI) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0, len(shareTypes))
	for _, s := range shareTypes {
		obj, d := cloudStorageFileShareTypeReferenceToObject(s)
		diags.Append(d...)
		items = append(items, obj)
	}
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: cloudStorageFileShareTypeReferenceAttrTypes()}), diags
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: cloudStorageFileShareTypeReferenceAttrTypes()}, items)
	diags.Append(d...)
	return list, diags
}

func (d *cloudStorageFileShareTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_storage_file_share_types"
}

func (d *cloudStorageFileShareTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if config := cloudReferenceConfigure(req, resp); config != nil {
		d.config = config
	}
}

func (d *cloudStorageFileShareTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to list the file storage types sold to a public cloud project, with their protocols, performance and size limits. This is read-only reference data: a file storage type sold in several regions appears once per region.",
		MarkdownDescription: "Use this data source to list the file storage types sold to a public cloud project, with their protocols, performance and size limits. This is read-only reference data: a file storage type sold in several regions appears once per region.",
		Attributes: map[string]schema.Attribute{
			"service_name": cloudReferenceServiceNameAttribute(),
			"region":       cloudReferenceFilterAttribute("Restrict the listing to the file storage types sold in this region."),
			"share_types": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "File storage types, one entry per type and region.",
				MarkdownDescription: "File storage types, one entry per type and region.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": cloudReferenceComputedString("File storage type identifier, to use as the type when creating a file storage"),
						"protocols": schema.ListAttribute{
							ElementType:         ovhtypes.TfStringType{},
							Computed:            true,
							Description:         "Protocols a file storage of this type supports",
							MarkdownDescription: "Protocols a file storage of this type supports",
						},
						"location": cloudReferenceLocationAttribute(
							"Region where the file storage type is sold to the project, whether that region is already activated on the project, and its availability zones",
						),
						"iops":         cloudReferencePerformanceAttribute("Input/output operations per second of a file storage of this type in this region"),
						"throughput":   cloudReferencePerformanceAttribute("Throughput of a file storage of this type in this region; null when no throughput figure is published for this region"),
						"minimum_size": cloudReferenceComputedInt64("Minimum size of a file storage of this type in this region, in GB"),
						"maximum_size": cloudReferenceComputedInt64("Maximum size of a file storage of this type in this region, in GB"),
					},
				},
			},
		},
	}
}

func (d *cloudStorageFileShareTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudStorageFileShareTypesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{}
	cloudReferenceSetQuery(query, "region", data.Region)
	endpoint := cloudReferenceEndpoint(data.ServiceName.ValueString(), "reference/storage/file/shareType", query)

	var shareTypes []CloudStorageFileShareTypeReferenceAPI
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &shareTypes); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error calling Get %s", endpoint), err.Error())
		return
	}

	list, diags := cloudStorageFileShareTypesToList(shareTypes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ShareTypes = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
