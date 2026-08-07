package v1beta1

import (
	"encoding/json"
	"testing"

	"k8s.io/utils/ptr"
)

// azureNodePoolPlatformNMinus1 represents the previous version of AzureNodePoolPlatform
// before securityType and uefiSettings were added.
type azureNodePoolPlatformNMinus1 struct {
	VMSize           string `json:"vmSize"`
	EncryptionAtHost string `json:"encryptionAtHost,omitempty"`
	SubnetID         string `json:"subnetID"`
}

func TestAzureNodePoolPlatformSerializationCompatibility(t *testing.T) {
	tests := []struct {
		name          string
		current       AzureNodePoolPlatform
		expectedJSON  string
		nMinus1Result azureNodePoolPlatformNMinus1
	}{
		{
			name: "When securityType and uefiSettings are omitted they should not appear in JSON",
			current: AzureNodePoolPlatform{
				VMSize:           "Standard_D2_v2",
				EncryptionAtHost: "Enabled",
				SubnetID:         "test-subnet",
			},
			expectedJSON: `{"vmSize":"Standard_D2_v2","image":{"type":""},"osDisk":{},"encryptionAtHost":"Enabled","subnetID":"test-subnet"}`,
			nMinus1Result: azureNodePoolPlatformNMinus1{
				VMSize:           "Standard_D2_v2",
				EncryptionAtHost: "Enabled",
				SubnetID:         "test-subnet",
			},
		},
		{
			name: "When securityType is set without uefiSettings it should serialize and N-1 should ignore it",
			current: AzureNodePoolPlatform{
				VMSize:           "Standard_D2_v2",
				SecurityType:     AzureSecurityTypeTrustedLaunch,
				EncryptionAtHost: "Enabled",
				SubnetID:         "test-subnet",
			},
			expectedJSON: `{"vmSize":"Standard_D2_v2","image":{"type":""},"osDisk":{},"encryptionAtHost":"Enabled","securityType":"TrustedLaunch","subnetID":"test-subnet"}`,
			nMinus1Result: azureNodePoolPlatformNMinus1{
				VMSize:           "Standard_D2_v2",
				EncryptionAtHost: "Enabled",
				SubnetID:         "test-subnet",
			},
		},
		{
			name: "When securityType and uefiSettings are both set N-1 should ignore them",
			current: AzureNodePoolPlatform{
				VMSize:       "Standard_D2_v2",
				SecurityType: AzureSecurityTypeConfidentialVM,
				UefiSettings: &AzureUefiSettings{
					SecureBoot: "Enabled",
					VTpm:       "Disabled",
				},
				SubnetID: "test-subnet",
			},
			expectedJSON: `{"vmSize":"Standard_D2_v2","image":{"type":""},"osDisk":{},"securityType":"ConfidentialVM","uefiSettings":{"secureBoot":"Enabled","vtpm":"Disabled"},"subnetID":"test-subnet"}`,
			nMinus1Result: azureNodePoolPlatformNMinus1{
				VMSize:   "Standard_D2_v2",
				SubnetID: "test-subnet",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.current)
			if err != nil {
				t.Fatalf("failed to marshal current struct: %v", err)
			}
			if string(data) != tt.expectedJSON {
				t.Errorf("unexpected JSON output:\n  got  %s\n  want %s", string(data), tt.expectedJSON)
			}

			var nMinus1 azureNodePoolPlatformNMinus1
			if err := json.Unmarshal(data, &nMinus1); err != nil {
				t.Fatalf("N-1 failed to unmarshal JSON from N: %v", err)
			}
			if nMinus1 != tt.nMinus1Result {
				t.Errorf("N-1 deserialization mismatch: got %+v, want %+v", nMinus1, tt.nMinus1Result)
			}

			nMinus1Data, err := json.Marshal(tt.nMinus1Result)
			if err != nil {
				t.Fatalf("failed to marshal N-1 struct: %v", err)
			}
			var roundTripped AzureNodePoolPlatform
			if err := json.Unmarshal(nMinus1Data, &roundTripped); err != nil {
				t.Fatalf("N failed to unmarshal JSON from N-1: %v", err)
			}
			if roundTripped.VMSize != tt.nMinus1Result.VMSize {
				t.Errorf("VMSize mismatch: got %s, want %s", roundTripped.VMSize, tt.nMinus1Result.VMSize)
			}
			if roundTripped.EncryptionAtHost != tt.nMinus1Result.EncryptionAtHost {
				t.Errorf("EncryptionAtHost mismatch: got %s, want %s", roundTripped.EncryptionAtHost, tt.nMinus1Result.EncryptionAtHost)
			}
			if roundTripped.SecurityType != "" {
				t.Errorf("SecurityType should be empty from N-1 data, got %s", roundTripped.SecurityType)
			}
			if roundTripped.UefiSettings != nil {
				t.Errorf("UefiSettings should be nil from N-1 data, got %+v", roundTripped.UefiSettings)
			}
		})
	}
}

// These types represent the N-1 (previous) version of the API structs,
// before the int32 -> *int32 pointer change. They are used to verify
// that JSON produced by the current types can be deserialized by
// previous versions of the code, and vice versa.
type nodePoolAutoScalingNMinus1 struct {
	Min int32 `json:"min"`
	Max int32 `json:"max"`
}

func TestNodePoolAutoScalingSerializationCompatibility(t *testing.T) {
	tests := []struct {
		name string
		// current is the N (current) version of the struct
		current NodePoolAutoScaling
		// expectedJSON is the expected JSON output from marshalling current
		expectedJSON string
		// nMinus1Result is the expected result when unmarshalling into the N-1 struct
		nMinus1Result nodePoolAutoScalingNMinus1
	}{
		{
			name: "When Min is set to a positive value it should round-trip to N-1",
			current: NodePoolAutoScaling{
				Min: ptr.To[int32](3),
				Max: 5,
			},
			expectedJSON:  `{"min":3,"max":5}`,
			nMinus1Result: nodePoolAutoScalingNMinus1{Min: 3, Max: 5},
		},
		{
			name: "When Min is explicitly zero it should round-trip to N-1",
			current: NodePoolAutoScaling{
				Min: ptr.To[int32](0),
				Max: 5,
			},
			expectedJSON:  `{"min":0,"max":5}`,
			nMinus1Result: nodePoolAutoScalingNMinus1{Min: 0, Max: 5},
		},
		{
			name: "When Min is nil it should be omitted and N-1 should deserialize as zero value",
			current: NodePoolAutoScaling{
				Min: nil,
				Max: 5,
			},
			expectedJSON:  `{"max":5}`,
			nMinus1Result: nodePoolAutoScalingNMinus1{Min: 0, Max: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal current (N) version
			data, err := json.Marshal(tt.current)
			if err != nil {
				t.Fatalf("failed to marshal current struct: %v", err)
			}
			if string(data) != tt.expectedJSON {
				t.Errorf("unexpected JSON output: got %s, want %s", string(data), tt.expectedJSON)
			}

			// Deserialize into N-1 struct
			var nMinus1 nodePoolAutoScalingNMinus1
			if err := json.Unmarshal(data, &nMinus1); err != nil {
				t.Fatalf("N-1 failed to unmarshal JSON from N: %v", err)
			}
			if nMinus1 != tt.nMinus1Result {
				t.Errorf("N-1 deserialization mismatch: got %+v, want %+v", nMinus1, tt.nMinus1Result)
			}

			// Reverse: marshal N-1 and deserialize into current (N)
			nMinus1Data, err := json.Marshal(tt.nMinus1Result)
			if err != nil {
				t.Fatalf("failed to marshal N-1 struct: %v", err)
			}
			var roundTripped NodePoolAutoScaling
			if err := json.Unmarshal(nMinus1Data, &roundTripped); err != nil {
				t.Fatalf("N failed to unmarshal JSON from N-1: %v", err)
			}
			if roundTripped.Max != tt.nMinus1Result.Max {
				t.Errorf("Max mismatch after N-1 round-trip: got %d, want %d", roundTripped.Max, tt.nMinus1Result.Max)
			}
			if ptr.Deref(roundTripped.Min, -1) != tt.nMinus1Result.Min {
				t.Errorf("Min mismatch after N-1 round-trip: got %v, want %d", roundTripped.Min, tt.nMinus1Result.Min)
			}
		})
	}
}
