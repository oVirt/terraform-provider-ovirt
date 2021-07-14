package ovirt_test

import (
	"fmt"
	"strings"
	"testing"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestParseResourceID(t *testing.T) {
	validResourceID := []struct {
		ResourceID string
		Count      int
	}{
		{
			"08f9d3ec-5768-479f-aa16-d2d9934b356a:8daea363-7535-4af3-a80f-e0f6c02666e0",
			2,
		},
		{
			"24252dc8-2a5a-4871-9601-7486a775d0e3:f59d3de3-512e-4eb3-9f4b-0e2395297b7c:8daea363-7535-4af3-a80f-e0f6c02666e0",
			3,
		},
		{
			"df18a81a-02eb-476e-bd42-bf87cbd5b948:615bdbb1-4d2a-443d-8ede-4f15b141603c:4104093f-5ba8-4d9f-849b-e4ad75143be7:7e8f03c7-0ac0-40b8-9497-9c64d4001f54",
			4,
		},
	}

	for _, v := range validResourceID {
		_, err := parseResourceID(v.ResourceID, v.Count)
		if err != nil {
			t.Fatalf("%s should be a valid Resource ID: %s", v.ResourceID, err)
		}
	}

	invalidResourceID := []struct {
		ResourceID string
		Count      int
	}{
		{
			"08f9d3ec-5768-479f-aa16-d2d9934b356a:8daea363-7535-4af3-a80f-e0f6c02666e0",
			3,
		},
		{
			"24252dc8-2a5a-4871-9601-7486a775d0e3:f59d3de3-512e-4eb3-9f4b-0e2395297b7c:8daea363-7535-4af3-a80f-e0f6c02666e0",
			2,
		},
		{
			"df18a81a-02eb-476e-bd42-bf87cbd5b948:615bdbb1-4d2a-443d-8ede-4f15b141603c:4104093f-5ba8-4d9f-849b-e4ad75143be7:7e8f03c7-0ac0-40b8-9497-9c64d4001f54",
			5,
		},
	}

	for _, v := range invalidResourceID {
		_, err := parseResourceID(v.ResourceID, v.Count)
		if err == nil {
			t.Fatalf("%s should be an invalid Resource ID: %s", v.ResourceID, err)
		}
	}

}

// Deprecated: this function should be moved to the test suite
func parseResourceID(id string, count int) ([]string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != count {
		return nil, fmt.Errorf("Invalid Resource ID %s, expected %d parts, got %d", id, count, len(parts))
	}
	return parts, nil
}

// Deprecated: this function should be moved to the test suite
func searchVmsByTag(service *ovirtsdk4.VmsService, tagName string) ([]string, error) {
	var vmIDs []string
	resp, err := service.List().Search(fmt.Sprintf("tag=%s", tagName)).Send()
	if err != nil {
		return nil, err
	}
	for _, v := range resp.MustVms().Slice() {
		vmIDs = append(vmIDs, v.MustId())
	}
	return vmIDs, nil
}

// Deprecated: this function should be moved to the test suite
func searchHostsByTag(service *ovirtsdk4.HostsService, tagName string) ([]string, error) {
	var hostIDs []string
	resp, err := service.List().Search(fmt.Sprintf("tag=%s", tagName)).Send()
	if err != nil {
		return nil, err
	}
	for _, v := range resp.MustHosts().Slice() {
		hostIDs = append(hostIDs, v.MustId())
	}
	return hostIDs, nil
}
