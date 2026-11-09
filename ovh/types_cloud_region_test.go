package ovh

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCloudRegionMergeWith_TypeAndActivated(t *testing.T) {
	cases := []struct {
		name          string
		body          string
		wantType      types.String
		wantActivated types.Bool
	}{
		{"3AZ activated", `{"name":"EU-WEST-PAR","type":"REGION-3-AZ","activated":true}`, types.StringValue("REGION-3-AZ"), types.BoolValue(true)},
		{"unknown type not activated", `{"name":"GRA11","type":null,"activated":false}`, types.StringNull(), types.BoolValue(false)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var region CloudRegion
			if err := json.Unmarshal([]byte(tc.body), &region); err != nil {
				t.Fatal(err)
			}

			var m cloudRegionDataSourceModel
			if diags := m.MergeWith(context.Background(), &region); diags.HasError() {
				t.Fatal(diags)
			}
			if !m.Type.StringValue.Equal(tc.wantType) {
				t.Errorf("type = %v, want %v", m.Type, tc.wantType)
			}
			if !m.Activated.Equal(tc.wantActivated) {
				t.Errorf("activated = %v, want %v", m.Activated, tc.wantActivated)
			}

			obj, diags := cloudRegionToObject(context.Background(), region)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if got := obj.Attributes()["activated"]; !got.Equal(tc.wantActivated) {
				t.Errorf("regions[].activated = %v, want %v", got, tc.wantActivated)
			}
		})
	}
}

func TestCloudRegionMergeWith_ObjectStorage(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantNull bool
	}{
		{
			"object storage region",
			`{"name":"GRA","objectStorage":{"bareContainerOnly":false,"isReplicationTarget":true,"lifecycleTargetClasses":["STANDARD_IA"],"replicationSupported":true,"storageClasses":["STANDARD","STANDARD_IA"]}}`,
			false,
		},
		{"no object storage", `{"name":"GRA11","objectStorage":null}`, true},
		{"absent", `{"name":"GRA11"}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var region CloudRegion
			if err := json.Unmarshal([]byte(tc.body), &region); err != nil {
				t.Fatal(err)
			}

			var m cloudRegionDataSourceModel
			if diags := m.MergeWith(context.Background(), &region); diags.HasError() {
				t.Fatal(diags)
			}
			obj, diags := cloudRegionToObject(context.Background(), region)
			if diags.HasError() {
				t.Fatal(diags)
			}
			listed, ok := obj.Attributes()["object_storage"].(types.Object)
			if !ok {
				t.Fatalf("regions[].object_storage missing or not an Object: %v", obj.Attributes())
			}

			wantType := types.ObjectType{AttrTypes: cloudRegionObjectStorageAttrTypes()}
			for label, got := range map[string]types.Object{"object_storage": m.ObjectStorage, "regions[].object_storage": listed} {
				if !got.Type(context.Background()).Equal(wantType) {
					t.Errorf("%s has the wrong type: %v", label, got.Type(context.Background()))
				}
				if got.IsNull() != tc.wantNull {
					t.Errorf("%s null = %v, want %v", label, got.IsNull(), tc.wantNull)
				}
			}
			if tc.wantNull {
				return
			}

			attrs := m.ObjectStorage.Attributes()
			if got := attrs["bare_container_only"]; !got.Equal(types.BoolValue(false)) {
				t.Errorf("bare_container_only = %v, want false", got)
			}
			if got := attrs["is_replication_target"]; !got.Equal(types.BoolValue(true)) {
				t.Errorf("is_replication_target = %v, want true", got)
			}
			if got := attrs["replication_supported"]; !got.Equal(types.BoolValue(true)) {
				t.Errorf("replication_supported = %v, want true", got)
			}
			wantLifecycle, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"STANDARD_IA"})
			if got := attrs["lifecycle_target_classes"]; !got.Equal(wantLifecycle) {
				t.Errorf("lifecycle_target_classes = %v, want %v", got, wantLifecycle)
			}
			wantClasses, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"STANDARD", "STANDARD_IA"})
			if got := attrs["storage_classes"]; !got.Equal(wantClasses) {
				t.Errorf("storage_classes = %v, want %v", got, wantClasses)
			}
			if !listed.Equal(m.ObjectStorage) {
				t.Errorf("regions[].object_storage = %v, want %v", listed, m.ObjectStorage)
			}
		})
	}
}
