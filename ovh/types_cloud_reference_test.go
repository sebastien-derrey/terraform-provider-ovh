package ovh

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

func referenceTestObjectAttr(t *testing.T, obj attr.Value, key string) attr.Value {
	t.Helper()
	o, ok := obj.(types.Object)
	if !ok {
		t.Fatalf("%v is not an object", obj)
	}
	v, ok := o.Attributes()[key]
	if !ok {
		t.Fatalf("attribute %q missing from %v", key, o)
	}
	return v
}

func referenceTestString(v string) ovhtypes.TfStringValue {
	return ovhtypes.TfStringValue{StringValue: types.StringValue(v)}
}

func TestCloudReferenceEndpoint(t *testing.T) {
	cases := []struct {
		name  string
		query url.Values
		want  string
	}{
		{"no filter", url.Values{}, "/v2/publicCloud/project/my%2Fproject/reference/gateway/model"},
		{"region and name", url.Values{"region": {"EU-WEST-PAR"}, "name": {"2XL"}}, "/v2/publicCloud/project/my%2Fproject/reference/gateway/model?name=2XL&region=EU-WEST-PAR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cloudReferenceEndpoint("my/project", "reference/gateway/model", tc.query); got != tc.want {
				t.Errorf("endpoint = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCloudReferenceSetQuery(t *testing.T) {
	query := url.Values{}
	cloudReferenceSetQuery(query, "region", ovhtypes.TfStringValue{StringValue: types.StringNull()})
	cloudReferenceSetQuery(query, "name", referenceTestString(""))
	if len(query) != 0 {
		t.Fatalf("null and empty filters must not be sent, got %v", query)
	}
	cloudReferenceSetQuery(query, "region", referenceTestString("GRA11"))
	if query.Get("region") != "GRA11" {
		t.Errorf("region = %q, want GRA11", query.Get("region"))
	}
}

func TestCloudLoadbalancerFlavorsToList(t *testing.T) {
	body := `[{"name":"SMALL","location":{"region":"EU-WEST-PAR","regionActivated":true,"availabilityZones":["eu-west-par-a","eu-west-par-b","eu-west-par-c"]},"bandwidth":200,"connectionsPerSecond":10000,"requestsPerSecond":2000,"price":{"value":1}},
	           {"name":"XL","location":{"region":"GRA11","regionActivated":false,"availabilityZones":[]},"bandwidth":4000,"connectionsPerSecond":80000,"requestsPerSecond":40000}]`
	var flavors []CloudLoadbalancerFlavorReferenceAPI
	if err := json.Unmarshal([]byte(body), &flavors); err != nil {
		t.Fatal(err)
	}
	list, diags := cloudLoadbalancerFlavorsToList(flavors)
	if diags.HasError() {
		t.Fatal(diags)
	}
	items := list.Elements()
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}

	first := items[0]
	if got := referenceTestObjectAttr(t, first, "name"); !got.Equal(referenceTestString("SMALL")) {
		t.Errorf("name = %v", got)
	}
	for key, want := range map[string]int64{"bandwidth": 200, "connections_per_second": 10000, "requests_per_second": 2000} {
		if got := referenceTestObjectAttr(t, first, key); !got.Equal(types.Int64Value(want)) {
			t.Errorf("%s = %v, want %d", key, got, want)
		}
	}
	location := referenceTestObjectAttr(t, first, "location")
	if got := referenceTestObjectAttr(t, location, "region"); !got.Equal(referenceTestString("EU-WEST-PAR")) {
		t.Errorf("location.region = %v", got)
	}
	if got := referenceTestObjectAttr(t, location, "region_activated"); !got.Equal(types.BoolValue(true)) {
		t.Errorf("location.region_activated = %v", got)
	}
	if got := referenceTestObjectAttr(t, location, "availability_zones").(types.List); len(got.Elements()) != 3 {
		t.Errorf("location.availability_zones = %v, want 3 zones", got)
	}

	second := referenceTestObjectAttr(t, items[1], "location")
	if got := referenceTestObjectAttr(t, second, "region_activated"); !got.Equal(types.BoolValue(false)) {
		t.Errorf("location.region_activated = %v, want false", got)
	}
	if got := referenceTestObjectAttr(t, second, "availability_zones").(types.List); got.IsNull() || len(got.Elements()) != 0 {
		t.Errorf("location.availability_zones = %v, want empty list", got)
	}
}

func TestCloudLoadbalancerFlavorsToList_Empty(t *testing.T) {
	list, diags := cloudLoadbalancerFlavorsToList(nil)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if list.IsNull() || len(list.Elements()) != 0 {
		t.Errorf("list = %v, want empty non-null list", list)
	}
}

func TestCloudGatewayModelsToList(t *testing.T) {
	body := `[{"name":"S","location":{"region":"GRA11","regionActivated":true,"availabilityZones":[]},"bandwidth":200}]`
	var models []CloudGatewayModelReferenceAPI
	if err := json.Unmarshal([]byte(body), &models); err != nil {
		t.Fatal(err)
	}
	list, diags := cloudGatewayModelsToList(models)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(list.Elements()) != 1 {
		t.Fatalf("len = %d, want 1", len(list.Elements()))
	}
	item := list.Elements()[0]
	if got := referenceTestObjectAttr(t, item, "name"); !got.Equal(referenceTestString("S")) {
		t.Errorf("name = %v", got)
	}
	if got := referenceTestObjectAttr(t, item, "bandwidth"); !got.Equal(types.Int64Value(200)) {
		t.Errorf("bandwidth = %v", got)
	}
	if got := referenceTestObjectAttr(t, referenceTestObjectAttr(t, item, "location"), "region"); !got.Equal(referenceTestString("GRA11")) {
		t.Errorf("location.region = %v", got)
	}
}

func TestCloudStorageBlockVolumeTypesToList(t *testing.T) {
	body := `[{"name":"HIGH_SPEED_GEN2","location":{"region":"GRA11","regionActivated":true,"availabilityZones":[]},"maximumSize":12000,
	            "iops":{"guaranteed":true,"level":{"value":30,"unit":"IOPS/GB"},"maximum":{"value":20000,"unit":"IOPS"}},
	            "throughput":{"guaranteed":false,"level":{"value":0.5,"unit":"MB/S/GB"},"maximum":{"value":1000,"unit":"MB/S"}}},
	           {"name":"CLASSIC","location":{"region":"GRA11","regionActivated":true,"availabilityZones":[]},"maximumSize":12000,
	            "iops":{"guaranteed":true,"level":{"value":250,"unit":"IOPS"},"maximum":{"value":250,"unit":"IOPS"}},
	            "throughput":null}]`
	var volumeTypes []CloudStorageBlockVolumeTypeReferenceAPI
	if err := json.Unmarshal([]byte(body), &volumeTypes); err != nil {
		t.Fatal(err)
	}
	list, diags := cloudStorageBlockVolumeTypesToList(volumeTypes)
	if diags.HasError() {
		t.Fatal(diags)
	}
	items := list.Elements()
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}

	first := items[0]
	if got := referenceTestObjectAttr(t, first, "maximum_size"); !got.Equal(types.Int64Value(12000)) {
		t.Errorf("maximum_size = %v", got)
	}
	iops := referenceTestObjectAttr(t, first, "iops")
	if got := referenceTestObjectAttr(t, iops, "guaranteed"); !got.Equal(types.BoolValue(true)) {
		t.Errorf("iops.guaranteed = %v", got)
	}
	level := referenceTestObjectAttr(t, iops, "level")
	if got := referenceTestObjectAttr(t, level, "value"); !got.Equal(types.Float64Value(30)) {
		t.Errorf("iops.level.value = %v", got)
	}
	if got := referenceTestObjectAttr(t, level, "unit"); !got.Equal(referenceTestString("IOPS/GB")) {
		t.Errorf("iops.level.unit = %v", got)
	}
	throughputMax := referenceTestObjectAttr(t, referenceTestObjectAttr(t, first, "throughput"), "maximum")
	if got := referenceTestObjectAttr(t, throughputMax, "value"); !got.Equal(types.Float64Value(1000)) {
		t.Errorf("throughput.maximum.value = %v", got)
	}

	if got := referenceTestObjectAttr(t, items[1], "throughput"); !got.IsNull() {
		t.Errorf("throughput = %v, want null", got)
	}
}

func TestCloudStorageFileShareTypesToList(t *testing.T) {
	body := `[{"name":"STANDARD_1AZ","protocols":["NFS"],"location":{"region":"GRA9","regionActivated":false,"availabilityZones":[]},"minimumSize":150,"maximumSize":10240,
	            "iops":{"guaranteed":false,"level":{"value":24,"unit":"IOPS/GB"},"maximum":{"value":16000,"unit":"IOPS"}},
	            "throughput":null}]`
	var shareTypes []CloudStorageFileShareTypeReferenceAPI
	if err := json.Unmarshal([]byte(body), &shareTypes); err != nil {
		t.Fatal(err)
	}
	list, diags := cloudStorageFileShareTypesToList(shareTypes)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(list.Elements()) != 1 {
		t.Fatalf("len = %d, want 1", len(list.Elements()))
	}
	item := list.Elements()[0]
	if got := referenceTestObjectAttr(t, item, "protocols"); !got.Equal(buildStringListFromAPI([]string{"NFS"})) {
		t.Errorf("protocols = %v", got)
	}
	if got := referenceTestObjectAttr(t, item, "minimum_size"); !got.Equal(types.Int64Value(150)) {
		t.Errorf("minimum_size = %v", got)
	}
	if got := referenceTestObjectAttr(t, item, "maximum_size"); !got.Equal(types.Int64Value(10240)) {
		t.Errorf("maximum_size = %v", got)
	}
	if got := referenceTestObjectAttr(t, referenceTestObjectAttr(t, item, "location"), "region_activated"); !got.Equal(types.BoolValue(false)) {
		t.Errorf("location.region_activated = %v", got)
	}
	if got := referenceTestObjectAttr(t, item, "throughput"); !got.IsNull() {
		t.Errorf("throughput = %v, want null", got)
	}
}
