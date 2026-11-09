package ovh

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// CloudProductAvailabilityAPIResponse is publicCloud.reference.ProductAvailability; links are dropped on purpose.
type CloudProductAvailabilityAPIResponse struct {
	Type   string                             `json:"type"`
	Models []CloudProductAvailabilityAPIModel `json:"models"`
}

type CloudProductAvailabilityAPIModel struct {
	Name      string                                `json:"name"`
	Locations []CloudProductAvailabilityAPILocation `json:"locations"`
}

type CloudProductAvailabilityAPILocation struct {
	Name              string            `json:"name"`
	RegionActivated   bool              `json:"regionActivated"`
	AvailabilityZones []string          `json:"availabilityZones"`
	StockStatus       map[string]string `json:"stockStatus"`
	QuotaStatus       string            `json:"quotaStatus"`
	QuotaLimitedBy    []string          `json:"quotaLimitedBy"`
}

type CloudProductAvailabilityModel struct {
	ServiceName    ovhtypes.TfStringValue `tfsdk:"service_name"`
	Type           ovhtypes.TfStringValue `tfsdk:"type"`
	Region         ovhtypes.TfStringValue `tfsdk:"region"`
	Name           ovhtypes.TfStringValue `tfsdk:"name"`
	NameContains   ovhtypes.TfStringValue `tfsdk:"name_contains"`
	Availabilities types.List             `tfsdk:"availabilities"`
}

func cloudProductAvailabilityItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":               ovhtypes.TfStringType{},
		"name":               ovhtypes.TfStringType{},
		"region":             ovhtypes.TfStringType{},
		"region_activated":   types.BoolType,
		"availability_zones": types.ListType{ElemType: ovhtypes.TfStringType{}},
		"stock_status":       types.MapType{ElemType: ovhtypes.TfStringType{}},
		"quota_status":       ovhtypes.TfStringType{},
		"quota_limited_by":   types.ListType{ElemType: ovhtypes.TfStringType{}},
	}
}

func cloudProductAvailabilityStockStatus(stock map[string]string) (types.Map, diag.Diagnostics) {
	// Empty rather than null so values() and lookup() work in HCL for every product type.
	elems := make(map[string]attr.Value, len(stock))
	for az, status := range stock {
		elems[az] = ovhtypes.TfStringValue{StringValue: types.StringValue(status)}
	}
	return types.MapValue(ovhtypes.TfStringType{}, elems)
}

// cloudProductAvailabilityFlatten emits one entry per (model, location), keeping the API order.
func cloudProductAvailabilityFlatten(groups []CloudProductAvailabilityAPIResponse) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	itemType := types.ObjectType{AttrTypes: cloudProductAvailabilityItemAttrTypes()}
	items := []attr.Value{}

	for _, group := range groups {
		for _, model := range group.Models {
			for _, location := range model.Locations {
				stockStatus, d := cloudProductAvailabilityStockStatus(location.StockStatus)
				diags.Append(d...)
				obj, d := types.ObjectValue(cloudProductAvailabilityItemAttrTypes(), map[string]attr.Value{
					"type":               ovhtypes.TfStringValue{StringValue: types.StringValue(group.Type)},
					"name":               ovhtypes.TfStringValue{StringValue: types.StringValue(model.Name)},
					"region":             ovhtypes.TfStringValue{StringValue: types.StringValue(location.Name)},
					"region_activated":   types.BoolValue(location.RegionActivated),
					"availability_zones": buildStringListFromAPI(location.AvailabilityZones),
					"stock_status":       stockStatus,
					"quota_status":       ovhtypes.TfStringValue{StringValue: types.StringValue(location.QuotaStatus)},
					"quota_limited_by":   buildStringListFromAPI(location.QuotaLimitedBy),
				})
				diags.Append(d...)
				if diags.HasError() {
					return types.ListNull(itemType), diags
				}
				items = append(items, obj)
			}
		}
	}

	list, d := types.ListValue(itemType, items)
	diags.Append(d...)
	return list, diags
}
