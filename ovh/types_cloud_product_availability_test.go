package ovh

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

const cloudProductAvailabilityTestBody = `[
  {
    "type": "INSTANCE",
    "models": [
      {
        "name": "b3-8",
        "links": [{"kind": "SPECS", "href": "/v2/publicCloud/project/p/reference/instance/flavor?name=b3-8"}],
        "locations": [
          {
            "name": "EU-WEST-PAR",
            "regionActivated": true,
            "availabilityZones": ["eu-west-par-a", "eu-west-par-b"],
            "stockStatus": {"eu-west-par-a": "AVAILABLE", "eu-west-par-b": "LOW"},
            "quotaStatus": "INSUFFICIENT",
            "quotaLimitedBy": ["CORES", "RAM"],
            "links": [{"kind": "QUOTA", "href": "/v2/publicCloud/project/p/quota/region/EU-WEST-PAR"}]
          },
          {
            "name": "GRA11",
            "regionActivated": false,
            "availabilityZones": [],
            "stockStatus": {"GRA11": "OUT_OF_STOCK"},
            "quotaStatus": "UNKNOWN",
            "quotaLimitedBy": [],
            "links": []
          }
        ]
      },
      {"name": "b3-16", "links": [], "locations": null}
    ]
  },
  {"type": "VOLUME", "models": null},
  {
    "type": "GATEWAY",
    "models": [
      {
        "name": "S",
        "links": [],
        "locations": [
          {"name": "GRA11", "regionActivated": true, "availabilityZones": null, "quotaStatus": "SUFFICIENT", "quotaLimitedBy": []},
          {"name": "SBG5", "regionActivated": true, "availabilityZones": null, "stockStatus": null, "quotaStatus": "SUFFICIENT", "quotaLimitedBy": []}
        ]
      },
      {"name": "M", "links": [], "locations": []}
    ]
  }
]`

func cloudProductAvailabilityTestString(t *testing.T, obj types.Object, key string) string {
	t.Helper()
	v, ok := obj.Attributes()[key].(ovhtypes.TfStringValue)
	if !ok {
		t.Fatalf("%s missing or not a TfStringValue: %v", key, obj.Attributes())
	}
	return v.ValueString()
}

func cloudProductAvailabilityTestStrings(t *testing.T, list attr.Value) []string {
	t.Helper()
	l, ok := list.(types.List)
	if !ok {
		t.Fatalf("not a List: %v", list)
	}
	out := make([]string, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		out = append(out, e.(ovhtypes.TfStringValue).ValueString())
	}
	return out
}

func TestCloudProductAvailabilityFlatten(t *testing.T) {
	var groups []CloudProductAvailabilityAPIResponse
	if err := json.Unmarshal([]byte(cloudProductAvailabilityTestBody), &groups); err != nil {
		t.Fatal(err)
	}

	list, diags := cloudProductAvailabilityFlatten(groups)
	if diags.HasError() {
		t.Fatal(diags)
	}

	want := []struct{ typ, name, region string }{
		{"INSTANCE", "b3-8", "EU-WEST-PAR"},
		{"INSTANCE", "b3-8", "GRA11"},
		{"GATEWAY", "S", "GRA11"},
		{"GATEWAY", "S", "SBG5"},
	}
	elems := list.Elements()
	if len(elems) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(elems), len(want), elems)
	}
	objs := make([]types.Object, len(elems))
	for i, e := range elems {
		objs[i] = e.(types.Object)
		if got := cloudProductAvailabilityTestString(t, objs[i], "type"); got != want[i].typ {
			t.Errorf("[%d].type = %q, want %q", i, got, want[i].typ)
		}
		if got := cloudProductAvailabilityTestString(t, objs[i], "name"); got != want[i].name {
			t.Errorf("[%d].name = %q, want %q", i, got, want[i].name)
		}
		if got := cloudProductAvailabilityTestString(t, objs[i], "region"); got != want[i].region {
			t.Errorf("[%d].region = %q, want %q", i, got, want[i].region)
		}
	}

	par := objs[0].Attributes()
	if !par["region_activated"].Equal(types.BoolValue(true)) {
		t.Errorf("[0].region_activated = %v, want true", par["region_activated"])
	}
	if got := cloudProductAvailabilityTestStrings(t, par["availability_zones"]); len(got) != 2 || got[0] != "eu-west-par-a" || got[1] != "eu-west-par-b" {
		t.Errorf("[0].availability_zones = %v", got)
	}
	stock := par["stock_status"].(types.Map).Elements()
	if len(stock) != 2 || stock["eu-west-par-a"].(ovhtypes.TfStringValue).ValueString() != "AVAILABLE" || stock["eu-west-par-b"].(ovhtypes.TfStringValue).ValueString() != "LOW" {
		t.Errorf("[0].stock_status = %v", stock)
	}
	if got := cloudProductAvailabilityTestString(t, objs[0], "quota_status"); got != "INSUFFICIENT" {
		t.Errorf("[0].quota_status = %q", got)
	}
	if got := cloudProductAvailabilityTestStrings(t, par["quota_limited_by"]); len(got) != 2 || got[0] != "CORES" || got[1] != "RAM" {
		t.Errorf("[0].quota_limited_by = %v", got)
	}

	gra := objs[1].Attributes()
	if !gra["region_activated"].Equal(types.BoolValue(false)) {
		t.Errorf("[1].region_activated = %v, want false", gra["region_activated"])
	}
	if azs := gra["availability_zones"].(types.List); azs.IsNull() || len(azs.Elements()) != 0 {
		t.Errorf("[1].availability_zones = %v, want empty non-null list", azs)
	}

	for i := 2; i <= 3; i++ {
		if stock := objs[i].Attributes()["stock_status"].(types.Map); stock.IsNull() || len(stock.Elements()) != 0 {
			t.Errorf("[%d].stock_status = %v, want empty non-null map", i, stock)
		}
	}

	gw := objs[2].Attributes()
	if !gw["availability_zones"].(types.List).IsNull() {
		t.Errorf("[2].availability_zones = %v, want null", gw["availability_zones"])
	}
}

func TestCloudProductAvailabilityFlatten_Empty(t *testing.T) {
	for _, groups := range [][]CloudProductAvailabilityAPIResponse{nil, {{Type: "VOLUME"}}} {
		list, diags := cloudProductAvailabilityFlatten(groups)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if list.IsNull() || len(list.Elements()) != 0 {
			t.Errorf("availabilities = %v, want empty non-null list", list)
		}
	}
}

func TestCloudProductAvailabilityEndpoint(t *testing.T) {
	str := func(s string) ovhtypes.TfStringValue {
		return ovhtypes.TfStringValue{StringValue: types.StringValue(s)}
	}
	null := ovhtypes.TfStringValue{StringValue: types.StringNull()}

	cases := []struct {
		name string
		data CloudProductAvailabilityModel
		want string
	}{
		{
			"no filter",
			CloudProductAvailabilityModel{ServiceName: str("p 1"), Type: null, Region: null, Name: null, NameContains: null},
			"/v2/publicCloud/project/p%201/reference/productAvailability",
		},
		{
			"every filter",
			CloudProductAvailabilityModel{ServiceName: str("p"), Type: str("INSTANCE"), Region: str("GRA11"), Name: str("b3-8"), NameContains: str("b3")},
			"/v2/publicCloud/project/p/reference/productAvailability?name=b3-8&nameContains=b3&region=GRA11&type=INSTANCE",
		},
		{
			"every filter, escaped service name and values",
			CloudProductAvailabilityModel{ServiceName: str("my/project ?#%"), Type: str("INSTANCE"), Region: str("GRA 11"), Name: str("b3&8"), NameContains: str("b3=+")},
			"/v2/publicCloud/project/my%2Fproject%20%3F%23%25/reference/productAvailability?name=b3%268&nameContains=b3%3D%2B&region=GRA+11&type=INSTANCE",
		},
		{
			"unknown filter ignored",
			CloudProductAvailabilityModel{ServiceName: str("p"), Type: ovhtypes.TfStringValue{StringValue: types.StringUnknown()}, Region: null, Name: null, NameContains: null},
			"/v2/publicCloud/project/p/reference/productAvailability",
		},
		{
			"empty filter ignored",
			CloudProductAvailabilityModel{ServiceName: str("p"), Type: str(""), Region: str("GRA11"), Name: null, NameContains: null},
			"/v2/publicCloud/project/p/reference/productAvailability?region=GRA11",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cloudProductAvailabilityEndpoint(&tc.data); got != tc.want {
				t.Errorf("endpoint = %q, want %q", got, tc.want)
			}
		})
	}
}
