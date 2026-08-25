package materialrequest

import (
	"math"
	"testing"

	"github.com/remiehneppo/material-management/types"
)

func TestMaterialsBelongToProfileValidatesCanonicalMaterial(t *testing.T) {
	profile := &types.MaterialsProfile{Estimate: types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{
		"oil": {Name: "oil", Unit: "l", Quantity: 10},
	}}}
	tests := []struct {
		name     string
		material types.Material
		want     bool
	}{
		{"valid", types.Material{Name: "oil", Unit: "l", Quantity: 2}, true},
		{"negative", types.Material{Name: "oil", Unit: "l", Quantity: -1}, false},
		{"zero", types.Material{Name: "oil", Unit: "l", Quantity: 0}, false},
		{"not finite", types.Material{Name: "oil", Unit: "l", Quantity: math.Inf(1)}, false},
		{"wrong name", types.Material{Name: "diesel", Unit: "l", Quantity: 1}, false},
		{"wrong unit", types.Material{Name: "oil", Unit: "kg", Quantity: 1}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requested := types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{"oil": test.material}}
			if got := MaterialsBelongToProfile(profile, requested); got != test.want {
				t.Fatalf("MaterialsBelongToProfile() = %v, want %v", got, test.want)
			}
		})
	}
	if MaterialsBelongToProfile(profile, types.MaterialsForEquipment{}) {
		t.Fatal("empty material selection was accepted")
	}
}
