package ovh

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

const (
	cloudReferenceServiceNameDescription = "Service name of the resource representing the id of the cloud project"
)

// cloudReferencePerformanceAPI is publicCloud.reference.storage.block.VolumeTypePerformance.
type cloudReferencePerformanceAPI struct {
	Guaranteed bool                              `json:"guaranteed"`
	Level      cloudReferencePerformanceValueAPI `json:"level"`
	Maximum    cloudReferencePerformanceValueAPI `json:"maximum"`
}

// cloudReferencePerformanceValueAPI is publicCloud.reference.storage.block.VolumeTypePerformanceValue.
type cloudReferencePerformanceValueAPI struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func cloudReferenceEndpoint(serviceName, apiPath string, query url.Values) string {
	endpoint := "/v2/publicCloud/project/" + url.PathEscape(serviceName) + "/" + apiPath
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	return endpoint
}

func cloudReferenceSetQuery(query url.Values, key string, value ovhtypes.TfStringValue) {
	if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
		query.Set(key, value.ValueString())
	}
}

func cloudReferenceConfigure(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *Config {
	if req.ProviderData == nil {
		return nil
	}
	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return nil
	}
	return config
}

func cloudReferenceServiceNameAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		CustomType:          ovhtypes.TfStringType{},
		Required:            true,
		Description:         cloudReferenceServiceNameDescription,
		MarkdownDescription: cloudReferenceServiceNameDescription,
	}
}

func cloudReferenceFilterAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		CustomType:          ovhtypes.TfStringType{},
		Optional:            true,
		Description:         description,
		MarkdownDescription: description,
	}
}

func cloudReferenceComputedString(description string) schema.StringAttribute {
	return schema.StringAttribute{
		CustomType:          ovhtypes.TfStringType{},
		Computed:            true,
		Description:         description,
		MarkdownDescription: description,
	}
}

func cloudReferenceComputedInt64(description string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Computed:            true,
		Description:         description,
		MarkdownDescription: description,
	}
}

func cloudReferenceLocationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"region":             ovhtypes.TfStringType{},
		"region_activated":   types.BoolType,
		"availability_zones": types.ListType{ElemType: ovhtypes.TfStringType{}},
	}
}

func cloudReferenceLocationAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Description:         description,
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"region": cloudReferenceComputedString("Region name, for example GRA11"),
			"region_activated": schema.BoolAttribute{
				Computed:            true,
				Description:         instanceFlavorLocationRegionActivatedDescription,
				MarkdownDescription: instanceFlavorLocationRegionActivatedDescription,
			},
			"availability_zones": schema.ListAttribute{
				ElementType:         ovhtypes.TfStringType{},
				Computed:            true,
				Description:         instanceFlavorLocationAZsDescription,
				MarkdownDescription: instanceFlavorLocationAZsDescription,
			},
		},
	}
}

func buildCloudReferenceLocationObject(l cloudReferenceAPILocation) types.Object {
	obj, _ := types.ObjectValue(cloudReferenceLocationAttrTypes(), map[string]attr.Value{
		"region":             ovhtypes.TfStringValue{StringValue: types.StringValue(l.Region)},
		"region_activated":   types.BoolPointerValue(l.RegionActivated),
		"availability_zones": buildStringListFromAPI(l.AvailabilityZones),
	})
	return obj
}

func cloudReferencePerformanceValueAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value": types.Float64Type,
		"unit":  ovhtypes.TfStringType{},
	}
}

func cloudReferencePerformanceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"guaranteed": types.BoolType,
		"level":      types.ObjectType{AttrTypes: cloudReferencePerformanceValueAttrTypes()},
		"maximum":    types.ObjectType{AttrTypes: cloudReferencePerformanceValueAttrTypes()},
	}
}

func cloudReferencePerformanceValueAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Description:         description,
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"value": schema.Float64Attribute{
				Computed:            true,
				Description:         "Numeric value, expressed in unit",
				MarkdownDescription: "Numeric value, expressed in `unit`",
			},
			"unit": cloudReferenceComputedString(`Unit of the value (IOPS, IOPS/GB, MB/S or MB/S/GB); units ending in "/GB" scale with the size`),
		},
	}
}

func cloudReferencePerformanceAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Description:         description,
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"guaranteed": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the baseline level is guaranteed",
				MarkdownDescription: "Whether the baseline level is guaranteed",
			},
			"level":   cloudReferencePerformanceValueAttribute("Baseline performance; may be expressed per GB of size, see its unit"),
			"maximum": cloudReferencePerformanceValueAttribute("Upper performance limit"),
		},
	}
}

func buildCloudReferencePerformanceValueObject(v cloudReferencePerformanceValueAPI) types.Object {
	obj, _ := types.ObjectValue(cloudReferencePerformanceValueAttrTypes(), map[string]attr.Value{
		"value": types.Float64Value(v.Value),
		"unit":  ovhtypes.TfStringValue{StringValue: types.StringValue(v.Unit)},
	})
	return obj
}

func buildCloudReferencePerformanceObject(p *cloudReferencePerformanceAPI) types.Object {
	if p == nil {
		return types.ObjectNull(cloudReferencePerformanceAttrTypes())
	}
	obj, _ := types.ObjectValue(cloudReferencePerformanceAttrTypes(), map[string]attr.Value{
		"guaranteed": types.BoolValue(p.Guaranteed),
		"level":      buildCloudReferencePerformanceValueObject(p.Level),
		"maximum":    buildCloudReferencePerformanceValueObject(p.Maximum),
	})
	return obj
}
