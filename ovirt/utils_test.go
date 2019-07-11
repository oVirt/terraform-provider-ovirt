package ovirt

import "testing"

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
