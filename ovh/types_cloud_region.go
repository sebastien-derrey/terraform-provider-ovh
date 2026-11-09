package ovh

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// CloudRegion is the API response for publicCloud.reference.Region (OVHcloud API v2),
// returned by GET /v2/publicCloud/project/{serviceName}/reference/region[/{name}].
type CloudRegion struct {
	Name              string                       `json:"name"`
	Status            string                       `json:"status"`
	Continent         string                       `json:"continent"`
	Country           string                       `json:"country"`
	DatacenterName    string                       `json:"datacenterName"`
	AvailabilityZones []string                     `json:"availabilityZones"`
	Services          []string                     `json:"services"`
	Type              *string                      `json:"type"`
	Activated         bool                         `json:"activated"`
	ObjectStorage     *cloudRegionAPIObjectStorage `json:"objectStorage"`
}

// cloudRegionAPIObjectStorage is publicCloud.reference.RegionObjectStorage.
type cloudRegionAPIObjectStorage struct {
	BareContainerOnly      bool     `json:"bareContainerOnly"`
	IsReplicationTarget    bool     `json:"isReplicationTarget"`
	LifecycleTargetClasses []string `json:"lifecycleTargetClasses"`
	ReplicationSupported   bool     `json:"replicationSupported"`
	StorageClasses         []string `json:"storageClasses"`
}

// cloudRegionDataSourceModel is the Terraform model for the singular ovh_cloud_region data source.
type cloudRegionDataSourceModel struct {
	ServiceName       ovhtypes.TfStringValue `tfsdk:"service_name"`
	Name              ovhtypes.TfStringValue `tfsdk:"name"`
	Status            ovhtypes.TfStringValue `tfsdk:"status"`
	Continent         ovhtypes.TfStringValue `tfsdk:"continent"`
	Country           ovhtypes.TfStringValue `tfsdk:"country"`
	DatacenterName    ovhtypes.TfStringValue `tfsdk:"datacenter_name"`
	AvailabilityZones types.List             `tfsdk:"availability_zones"`
	Services          types.List             `tfsdk:"services"`
	Type              ovhtypes.TfStringValue `tfsdk:"type"`
	Activated         types.Bool             `tfsdk:"activated"`
	ObjectStorage     types.Object           `tfsdk:"object_storage"`
}

// cloudRegionsDataSourceModel is the Terraform model for the plural ovh_cloud_regions data source.
type cloudRegionsDataSourceModel struct {
	ServiceName ovhtypes.TfStringValue `tfsdk:"service_name"`
	Regions     types.List             `tfsdk:"regions"`
}

// cloudRegionDetailAttributes returns the computed schema attributes describing a single
// region. It is shared between the singular data source and the plural data source's nested
// "regions" objects. The "name" attribute is intentionally excluded so each data source can
// declare it as required (singular input) or computed (plural list item).
func cloudRegionDetailAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"status": schema.StringAttribute{
			CustomType:          ovhtypes.TfStringType{},
			Computed:            true,
			Description:         "Region status (ENABLED, DISABLED or MAINTENANCE)",
			MarkdownDescription: "Region status (ENABLED, DISABLED or MAINTENANCE)",
		},
		"continent": schema.StringAttribute{
			CustomType:          ovhtypes.TfStringType{},
			Computed:            true,
			Description:         "Continent code of the region",
			MarkdownDescription: "Continent code of the region",
		},
		"country": schema.StringAttribute{
			CustomType:          ovhtypes.TfStringType{},
			Computed:            true,
			Description:         "Country code of the region",
			MarkdownDescription: "Country code of the region",
		},
		"datacenter_name": schema.StringAttribute{
			CustomType:          ovhtypes.TfStringType{},
			Computed:            true,
			Description:         "Display name of the datacenter hosting the region",
			MarkdownDescription: "Display name of the datacenter hosting the region",
		},
		"availability_zones": schema.ListAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			Description:         "Availability zones available in the region",
			MarkdownDescription: "Availability zones available in the region",
		},
		"services": schema.ListAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			Description:         "Available OpenStack services in the region",
			MarkdownDescription: "Available OpenStack services in the region",
		},
		"type": schema.StringAttribute{
			CustomType:          ovhtypes.TfStringType{},
			Computed:            true,
			Description:         "Region type (LOCAL-ZONE, REGION-1-AZ or REGION-3-AZ); null when the catalog reports an unrecognized type",
			MarkdownDescription: "Region type (`LOCAL-ZONE`, `REGION-1-AZ` or `REGION-3-AZ`); null when the catalog reports an unrecognized type",
		},
		"activated": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the region is activated on the project. When false, the region exists but is not yet activated for the project",
			MarkdownDescription: "Whether the region is activated on the project. When false, the region exists but is not yet activated for the project",
		},
		"object_storage": schema.SingleNestedAttribute{
			Computed:            true,
			Description:         "Object storage capabilities of the region; null when the region offers no object storage",
			MarkdownDescription: "Object storage capabilities of the region; null when the region offers no object storage",
			Attributes: map[string]schema.Attribute{
				"bare_container_only": schema.BoolAttribute{
					Computed:            true,
					Description:         "Whether the region only supports bare containers, without the additional bucket features (encryption, versioning, object lock, lifecycle, replication)",
					MarkdownDescription: "Whether the region only supports bare containers, without the additional bucket features (encryption, versioning, object lock, lifecycle, replication)",
				},
				"is_replication_target": schema.BoolAttribute{
					Computed:            true,
					Description:         "Whether a bucket in this region can be used as the destination of a replication rule",
					MarkdownDescription: "Whether a bucket in this region can be used as the destination of a replication rule",
				},
				"lifecycle_target_classes": schema.ListAttribute{
					ElementType:         types.StringType,
					Computed:            true,
					Description:         "Storage classes that can be used as a transition target in a lifecycle rule in this region",
					MarkdownDescription: "Storage classes that can be used as a transition target in a lifecycle rule in this region",
				},
				"replication_supported": schema.BoolAttribute{
					Computed:            true,
					Description:         "Whether replication rules can be configured on buckets located in this region",
					MarkdownDescription: "Whether replication rules can be configured on buckets located in this region",
				},
				"storage_classes": schema.ListAttribute{
					ElementType:         types.StringType,
					Computed:            true,
					Description:         "Storage classes orderable in this region",
					MarkdownDescription: "Storage classes orderable in this region",
				},
			},
		},
	}
}

// cloudRegionObjectAttrTypes returns the attribute types of a single region object, matching
// the attributes declared in cloudRegionDetailAttributes plus "name".
func cloudRegionObjectAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               ovhtypes.TfStringType{},
		"status":             ovhtypes.TfStringType{},
		"continent":          ovhtypes.TfStringType{},
		"country":            ovhtypes.TfStringType{},
		"datacenter_name":    ovhtypes.TfStringType{},
		"availability_zones": types.ListType{ElemType: types.StringType},
		"services":           types.ListType{ElemType: types.StringType},
		"type":               ovhtypes.TfStringType{},
		"activated":          types.BoolType,
		"object_storage":     types.ObjectType{AttrTypes: cloudRegionObjectStorageAttrTypes()},
	}
}

func cloudRegionObjectStorageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"bare_container_only":      types.BoolType,
		"is_replication_target":    types.BoolType,
		"lifecycle_target_classes": types.ListType{ElemType: types.StringType},
		"replication_supported":    types.BoolType,
		"storage_classes":          types.ListType{ElemType: types.StringType},
	}
}

func cloudRegionObjectStorageToObject(ctx context.Context, objectStorage *cloudRegionAPIObjectStorage) (types.Object, diag.Diagnostics) {
	if objectStorage == nil {
		return types.ObjectNull(cloudRegionObjectStorageAttrTypes()), nil
	}

	var diags diag.Diagnostics

	lifecycleTargetClasses, d := stringSliceToTfList(ctx, objectStorage.LifecycleTargetClasses)
	diags.Append(d...)

	storageClasses, d := stringSliceToTfList(ctx, objectStorage.StorageClasses)
	diags.Append(d...)

	obj, d := types.ObjectValue(cloudRegionObjectStorageAttrTypes(), map[string]attr.Value{
		"bare_container_only":      types.BoolValue(objectStorage.BareContainerOnly),
		"is_replication_target":    types.BoolValue(objectStorage.IsReplicationTarget),
		"lifecycle_target_classes": lifecycleTargetClasses,
		"replication_supported":    types.BoolValue(objectStorage.ReplicationSupported),
		"storage_classes":          storageClasses,
	})
	diags.Append(d...)

	return obj, diags
}

// stringSliceToTfList converts a Go string slice into a Terraform list of strings. A nil slice
// becomes a null list, while an empty (but non-nil) slice becomes an empty list.
func stringSliceToTfList(ctx context.Context, in []string) (types.List, diag.Diagnostics) {
	if in == nil {
		return types.ListNull(types.StringType), nil
	}
	return types.ListValueFrom(ctx, types.StringType, in)
}

// MergeWith populates the singular data source model from an API region.
func (m *cloudRegionDataSourceModel) MergeWith(ctx context.Context, region *CloudRegion) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = ovhtypes.TfStringValue{StringValue: types.StringValue(region.Name)}
	m.Status = ovhtypes.TfStringValue{StringValue: types.StringValue(region.Status)}
	m.Continent = ovhtypes.TfStringValue{StringValue: types.StringValue(region.Continent)}
	m.Country = ovhtypes.TfStringValue{StringValue: types.StringValue(region.Country)}
	m.DatacenterName = ovhtypes.TfStringValue{StringValue: types.StringValue(region.DatacenterName)}

	availabilityZones, d := stringSliceToTfList(ctx, region.AvailabilityZones)
	diags.Append(d...)
	m.AvailabilityZones = availabilityZones

	services, d := stringSliceToTfList(ctx, region.Services)
	diags.Append(d...)
	m.Services = services

	m.Type = ovhtypes.TfStringValue{StringValue: types.StringPointerValue(region.Type)}
	m.Activated = types.BoolValue(region.Activated)

	objectStorage, d := cloudRegionObjectStorageToObject(ctx, region.ObjectStorage)
	diags.Append(d...)
	m.ObjectStorage = objectStorage

	return diags
}

// cloudRegionToObject converts an API region into a Terraform object value matching
// cloudRegionObjectAttrTypes.
func cloudRegionToObject(ctx context.Context, region CloudRegion) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	availabilityZones, d := stringSliceToTfList(ctx, region.AvailabilityZones)
	diags.Append(d...)

	services, d := stringSliceToTfList(ctx, region.Services)
	diags.Append(d...)

	objectStorage, d := cloudRegionObjectStorageToObject(ctx, region.ObjectStorage)
	diags.Append(d...)

	obj, d := types.ObjectValue(cloudRegionObjectAttrTypes(), map[string]attr.Value{
		"name":               ovhtypes.TfStringValue{StringValue: types.StringValue(region.Name)},
		"status":             ovhtypes.TfStringValue{StringValue: types.StringValue(region.Status)},
		"continent":          ovhtypes.TfStringValue{StringValue: types.StringValue(region.Continent)},
		"country":            ovhtypes.TfStringValue{StringValue: types.StringValue(region.Country)},
		"datacenter_name":    ovhtypes.TfStringValue{StringValue: types.StringValue(region.DatacenterName)},
		"availability_zones": availabilityZones,
		"services":           services,
		"type":               ovhtypes.TfStringValue{StringValue: types.StringPointerValue(region.Type)},
		"activated":          types.BoolValue(region.Activated),
		"object_storage":     objectStorage,
	})
	diags.Append(d...)

	return obj, diags
}
