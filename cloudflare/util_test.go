package cloudflare

import (
	"github.com/ViRb3/wgcf/v2/openapi"
	"testing"
)

func TestFindDeviceReturnsPointerToSliceElement(t *testing.T) {
	name := "old"
	devices := []BoundDevice{
		BoundDevice(openapi.GetBoundDevices200Response{
			Activated: "a",
			Active:    true,
			Created:   "c",
			Id:        "id1",
			Model:     "m",
			Role:      "r",
			Type:      "t",
			Name:      &name,
		}),
	}
	d, err := FindDevice(devices, "id1")
	if err != nil {
		t.Fatalf("FindDevice error: %v", err)
	}
	newName := "new"
	d.Name = &newName
	if devices[0].Name == nil || *devices[0].Name != newName {
		t.Fatalf("expected slice element name %q, got %v", newName, devices[0].Name)
	}
}
