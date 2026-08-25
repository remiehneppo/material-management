package materialrequest

import (
	"testing"

	"github.com/remiehneppo/material-management/types"
)

func TestAccumulateAddsRealityWithoutMutatingUnrelatedMaterials(t *testing.T) {
	current := types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{
		"oil": {Name: "oil", Unit: "l", Quantity: 2},
	}}
	delta := types.MaterialsForEquipment{
		ConsumableSupplies:   map[string]types.Material{"oil": {Name: "oil", Unit: "l", Quantity: 3}},
		ReplacementMaterials: map[string]types.Material{"seal": {Name: "seal", Unit: "piece", Quantity: 1}},
	}

	got := Accumulate(current, delta)
	if got.ConsumableSupplies["oil"].Quantity != 5 {
		t.Fatalf("oil quantity = %v, want 5", got.ConsumableSupplies["oil"].Quantity)
	}
	if got.ReplacementMaterials["seal"].Quantity != 1 {
		t.Fatalf("seal quantity = %v, want 1", got.ReplacementMaterials["seal"].Quantity)
	}
}
