package ovh

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

func TestCloudInstanceFlavorMergeWith_LocationRegionActivated(t *testing.T) {
	cases := []struct {
		name string
		body string
		want types.Bool
	}{
		{"activated", `{"id":"f1","location":{"region":"GRA11","regionActivated":true}}`, types.BoolValue(true)},
		{"not activated", `{"id":"f1","location":{"region":"GRA11","regionActivated":false}}`, types.BoolValue(false)},
		{"absent", `{"id":"f1","location":{"region":"GRA11"}}`, types.BoolNull()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp CloudInstanceFlavorAPIResponse
			if err := json.Unmarshal([]byte(tc.body), &resp); err != nil {
				t.Fatal(err)
			}
			var m CloudInstanceFlavorModel
			m.MergeWith(context.Background(), &resp)

			got, ok := m.Location.Attributes()["region_activated"].(types.Bool)
			if !ok {
				t.Fatalf("region_activated missing or not a Bool: %v", m.Location.Attributes())
			}
			if !got.Equal(tc.want) {
				t.Errorf("region_activated = %v, want %v", got, tc.want)
			}
			if region := m.Location.Attributes()["region"]; region == nil || region.String() != `"GRA11"` {
				t.Errorf("region = %v, want GRA11", region)
			}
			if az := m.Location.Attributes()["availability_zone"]; az == nil || az.String() != `""` {
				t.Errorf("availability_zone = %v, want empty", az)
			}
		})
	}
}

func TestCloudInstanceFlavorMergeWith_LocationAvailabilityZones(t *testing.T) {
	cases := []struct {
		name string
		body string
		want types.List
	}{
		{
			"3AZ region",
			`{"id":"f1","location":{"region":"EU-WEST-PAR","availabilityZones":["eu-west-par-a","eu-west-par-b","eu-west-par-c"]}}`,
			buildStringListFromAPI([]string{"eu-west-par-a", "eu-west-par-b", "eu-west-par-c"}),
		},
		{"single-AZ region", `{"id":"f1","location":{"region":"GRA11","availabilityZones":[]}}`, buildStringListFromAPI([]string{})},
		{"absent", `{"id":"f1","location":{"region":"GRA11"}}`, types.ListNull(ovhtypes.TfStringType{})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp CloudInstanceFlavorAPIResponse
			if err := json.Unmarshal([]byte(tc.body), &resp); err != nil {
				t.Fatal(err)
			}
			var m CloudInstanceFlavorModel
			m.MergeWith(context.Background(), &resp)

			got, ok := m.Location.Attributes()["availability_zones"].(types.List)
			if !ok {
				t.Fatalf("availability_zones missing or not a List: %v", m.Location.Attributes())
			}
			if !got.Equal(tc.want) {
				t.Errorf("availability_zones = %v, want %v", got, tc.want)
			}
			if got.IsNull() != tc.want.IsNull() || len(got.Elements()) != len(tc.want.Elements()) {
				t.Errorf("availability_zones null=%v len=%d, want null=%v len=%d", got.IsNull(), len(got.Elements()), tc.want.IsNull(), len(tc.want.Elements()))
			}
		})
	}
}

func TestCloudInstanceFlavorMergeWith_NilLocation(t *testing.T) {
	var m CloudInstanceFlavorModel
	m.MergeWith(context.Background(), &CloudInstanceFlavorAPIResponse{Id: "f1"})
	if !m.Location.IsNull() {
		t.Errorf("location = %v, want null", m.Location)
	}
	if !m.Location.Type(context.Background()).Equal(types.ObjectType{AttrTypes: instanceFlavorLocationAttrTypes()}) {
		t.Errorf("null location has the wrong type: %v", m.Location.Type(context.Background()))
	}
}

func TestCloudInstanceFlavorMergeWith_GpusAndBandwidth(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		wantGpus       types.Int64
		wantPrivate    types.Int64
		wantInbound    types.Int64
		wantOutbound   types.Int64
		wantBandwidthN bool
	}{
		{
			"gpu flavor with full bandwidth",
			`{"id":"f1","gpus":2,"bandwidth":{"private":4000,"publicInbound":10000,"publicOutbound":2000}}`,
			types.Int64Value(2), types.Int64Value(4000), types.Int64Value(10000), types.Int64Value(2000), false,
		},
		{
			"undeclared bandwidth values",
			`{"id":"f1","gpus":0,"bandwidth":{"private":null,"publicInbound":null,"publicOutbound":500}}`,
			types.Int64Value(0), types.Int64Null(), types.Int64Null(), types.Int64Value(500), false,
		},
		{"absent bandwidth", `{"id":"f1"}`, types.Int64Value(0), types.Int64Null(), types.Int64Null(), types.Int64Null(), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp CloudInstanceFlavorAPIResponse
			if err := json.Unmarshal([]byte(tc.body), &resp); err != nil {
				t.Fatal(err)
			}
			var m CloudInstanceFlavorModel
			m.MergeWith(context.Background(), &resp)

			if !m.Gpus.Equal(tc.wantGpus) {
				t.Errorf("gpus = %v, want %v", m.Gpus, tc.wantGpus)
			}
			if !m.Bandwidth.Type(context.Background()).Equal(types.ObjectType{AttrTypes: instanceFlavorBandwidthAttrTypes()}) {
				t.Fatalf("bandwidth has the wrong type: %v", m.Bandwidth.Type(context.Background()))
			}
			if m.Bandwidth.IsNull() != tc.wantBandwidthN {
				t.Fatalf("bandwidth null = %v, want %v", m.Bandwidth.IsNull(), tc.wantBandwidthN)
			}
			if tc.wantBandwidthN {
				return
			}
			attrs := m.Bandwidth.Attributes()
			for key, want := range map[string]types.Int64{
				"private":         tc.wantPrivate,
				"public_inbound":  tc.wantInbound,
				"public_outbound": tc.wantOutbound,
			} {
				if got := attrs[key]; got == nil || !got.Equal(want) {
					t.Errorf("bandwidth.%s = %v, want %v", key, got, want)
				}
			}
		})
	}
}
