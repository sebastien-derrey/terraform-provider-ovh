package ovh

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// cloudInstanceAPILocation is publicCloud.common.Location (image catalog).
type cloudInstanceAPILocation struct {
	Region           string `json:"region"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

// cloudReferenceAPILocation is publicCloud.reference.Location (flavor catalog).
type cloudReferenceAPILocation struct {
	Region            string   `json:"region"`
	RegionActivated   *bool    `json:"regionActivated,omitempty"`
	AvailabilityZones []string `json:"availabilityZones"`
}

// cloudInstanceFlavorAPIBandwidth is publicCloud.instance.FlavorBandwidth, in Mbit/s.
type cloudInstanceFlavorAPIBandwidth struct {
	Private        *int64 `json:"private"`
	PublicInbound  *int64 `json:"publicInbound"`
	PublicOutbound *int64 `json:"publicOutbound"`
}

type CloudInstanceFlavorAPIResponse struct {
	Id          string                           `json:"id"`
	Name        string                           `json:"name"`
	Vcpus       int64                            `json:"vcpus"`
	Ram         int64                            `json:"ram"`
	Disk        int64                            `json:"disk"`
	Swap        int64                            `json:"swap"`
	Ephemeral   int64                            `json:"ephemeral"`
	Gpus        int64                            `json:"gpus"`
	IsPublic    bool                             `json:"isPublic"`
	Description string                           `json:"description"`
	Location    *cloudReferenceAPILocation       `json:"location,omitempty"`
	Bandwidth   *cloudInstanceFlavorAPIBandwidth `json:"bandwidth,omitempty"`
}

type CloudInstanceFlavorModel struct {
	ServiceName ovhtypes.TfStringValue `tfsdk:"service_name"`
	Id          ovhtypes.TfStringValue `tfsdk:"id"`
	Name        ovhtypes.TfStringValue `tfsdk:"name"`
	Vcpus       types.Int64            `tfsdk:"vcpus"`
	Ram         types.Int64            `tfsdk:"ram"`
	Disk        types.Int64            `tfsdk:"disk"`
	Swap        types.Int64            `tfsdk:"swap"`
	Ephemeral   types.Int64            `tfsdk:"ephemeral"`
	Gpus        types.Int64            `tfsdk:"gpus"`
	IsPublic    types.Bool             `tfsdk:"is_public"`
	Description ovhtypes.TfStringValue `tfsdk:"description"`
	Location    types.Object           `tfsdk:"location"`
	Bandwidth   types.Object           `tfsdk:"bandwidth"`
}

func instanceFlavorItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          ovhtypes.TfStringType{},
		"name":        ovhtypes.TfStringType{},
		"vcpus":       types.Int64Type,
		"ram":         types.Int64Type,
		"disk":        types.Int64Type,
		"swap":        types.Int64Type,
		"ephemeral":   types.Int64Type,
		"gpus":        types.Int64Type,
		"is_public":   types.BoolType,
		"description": ovhtypes.TfStringType{},
		"location":    types.ObjectType{AttrTypes: instanceFlavorLocationAttrTypes()},
		"bandwidth":   types.ObjectType{AttrTypes: instanceFlavorBandwidthAttrTypes()},
	}
}

func instanceFlavorBandwidthAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"private":         types.Int64Type,
		"public_inbound":  types.Int64Type,
		"public_outbound": types.Int64Type,
	}
}

func buildCloudInstanceFlavorBandwidthObject(b *cloudInstanceFlavorAPIBandwidth) types.Object {
	if b == nil {
		return types.ObjectNull(instanceFlavorBandwidthAttrTypes())
	}
	obj, _ := types.ObjectValue(instanceFlavorBandwidthAttrTypes(), map[string]attr.Value{
		"private":         types.Int64PointerValue(b.Private),
		"public_inbound":  types.Int64PointerValue(b.PublicInbound),
		"public_outbound": types.Int64PointerValue(b.PublicOutbound),
	})
	return obj
}

func instanceFlavorLocationAttrTypes() map[string]attr.Type {
	attrTypes := instanceLocationAttrTypes()
	attrTypes["region_activated"] = types.BoolType
	attrTypes["availability_zones"] = types.ListType{ElemType: ovhtypes.TfStringType{}}
	return attrTypes
}

// availability_zone is deprecated and always empty: publicCloud.reference.Location has no single zone.
func buildCloudInstanceFlavorLocationObject(l *cloudReferenceAPILocation) types.Object {
	if l == nil {
		return types.ObjectNull(instanceFlavorLocationAttrTypes())
	}
	obj, _ := types.ObjectValue(instanceFlavorLocationAttrTypes(), map[string]attr.Value{
		"region":             ovhtypes.TfStringValue{StringValue: types.StringValue(l.Region)},
		"availability_zone":  ovhtypes.TfStringValue{StringValue: types.StringValue("")},
		"region_activated":   types.BoolPointerValue(l.RegionActivated),
		"availability_zones": buildStringListFromAPI(l.AvailabilityZones),
	})
	return obj
}

func buildCloudInstanceLocationObject(l *cloudInstanceAPILocation) types.Object {
	if l == nil {
		return types.ObjectNull(instanceLocationAttrTypes())
	}
	obj, _ := types.ObjectValue(instanceLocationAttrTypes(), map[string]attr.Value{
		"region":            ovhtypes.TfStringValue{StringValue: types.StringValue(l.Region)},
		"availability_zone": ovhtypes.TfStringValue{StringValue: types.StringValue(l.AvailabilityZone)},
	})
	return obj
}

func (m *CloudInstanceFlavorModel) MergeWith(ctx context.Context, response *CloudInstanceFlavorAPIResponse) {
	m.Id = ovhtypes.TfStringValue{StringValue: types.StringValue(response.Id)}
	m.Name = ovhtypes.TfStringValue{StringValue: types.StringValue(response.Name)}
	m.Vcpus = types.Int64Value(response.Vcpus)
	m.Ram = types.Int64Value(response.Ram)
	m.Disk = types.Int64Value(response.Disk)
	m.Swap = types.Int64Value(response.Swap)
	m.Ephemeral = types.Int64Value(response.Ephemeral)
	m.Gpus = types.Int64Value(response.Gpus)
	m.IsPublic = types.BoolValue(response.IsPublic)
	m.Description = ovhtypes.TfStringValue{StringValue: types.StringValue(response.Description)}
	m.Location = buildCloudInstanceFlavorLocationObject(response.Location)
	m.Bandwidth = buildCloudInstanceFlavorBandwidthObject(response.Bandwidth)
}
