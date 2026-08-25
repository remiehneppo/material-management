package materialrequest

import (
	"math"

	"github.com/remiehneppo/material-management/types"
)

func MaterialsBelongToProfile(profile *types.MaterialsProfile, requested types.MaterialsForEquipment) bool {
	if len(requested.ConsumableSupplies)+len(requested.ReplacementMaterials) == 0 {
		return false
	}
	return materialMapBelongsToProfile(profile.Estimate.ConsumableSupplies, profile.Reality.ConsumableSupplies, requested.ConsumableSupplies) &&
		materialMapBelongsToProfile(profile.Estimate.ReplacementMaterials, profile.Reality.ReplacementMaterials, requested.ReplacementMaterials)
}

func materialMapBelongsToProfile(estimate, reality, requested map[string]types.Material) bool {
	for key, material := range requested {
		canonical, ok := estimate[key]
		if !ok {
			canonical, ok = reality[key]
		}
		if !ok || material.Name != key || canonical.Name != key || material.Unit != canonical.Unit ||
			material.Quantity <= 0 || math.IsNaN(material.Quantity) || math.IsInf(material.Quantity, 0) {
			return false
		}
	}
	return true
}
