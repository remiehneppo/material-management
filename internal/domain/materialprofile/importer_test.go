package materialprofile

import (
	"testing"

	"github.com/remiehneppo/material-management/types"
)

func TestParseRowsKeysProfilesByIndexPath(t *testing.T) {
	rows := [][]string{{"1.1", "Pump"}, {"", types.LABEL_CONSUMABLE}, {"-", "Oil", "L", "2"}, {"1.2", "Pump"}, {"", types.LABEL_REPLACEMENT}, {"-", "Seal", "piece", "1"}}
	profiles, err := parseRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 || profiles[0].Index == profiles[1].Index {
		t.Fatalf("profiles=%+v", profiles)
	}
}

func TestParseRowsRejectsDuplicateIndexPathBeforeWrites(t *testing.T) {
	_, err := parseRows([][]string{{"1.1", "Pump"}, {"1.1", "Other"}})
	if err == nil {
		t.Fatal("duplicate index path accepted")
	}
}

func TestMergeEstimatePreservesMaterialsAbsentFromPatch(t *testing.T) {
	current := types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{"old": {Name: "old", Quantity: 1}}}
	patch := types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{"new": {Name: "new", Quantity: 2}}}
	merged := mergeEstimate(current, patch)
	if len(merged.ConsumableSupplies) != 2 {
		t.Fatalf("merged=%+v", merged)
	}
}
