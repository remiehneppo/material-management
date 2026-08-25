package materialprofile

import (
	"context"
	"errors"

	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrInvalidMaterialCategory = errors.New("nhóm vật tư không hợp lệ")
	ErrConcurrentUpdate        = errors.New("hồ sơ vật tư vừa được thay đổi; vui lòng thử lại")
)

type UpsertMaterialRequest struct {
	Category string         `json:"category" binding:"required"`
	Material types.Material `json:"material" binding:"required"`
}

type Catalog struct{ db *mongo.Database }

func NewCatalog(db *mongo.Database) *Catalog { return &Catalog{db: db} }

func (c *Catalog) EnsureIndexes(ctx context.Context) error {
	_, err := c.db.Collection("materials_profiles").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "maintenance_instance_id", Value: 1}, {Key: "sector", Value: 1}, {Key: "index", Value: 1}},
		Options: options.Index().SetName("material_profile_business_key").SetUnique(true),
	})
	return err
}

func (c *Catalog) UpsertEstimatedMaterial(ctx context.Context, profileID string, request UpsertMaterialRequest) error {
	if request.Category != "consumable_supplies" && request.Category != "replacement_materials" {
		return ErrInvalidMaterialCategory
	}
	if request.Material.Name == "" || request.Material.Unit == "" || request.Material.Quantity < 0 {
		return errors.New("vui lòng nhập tên, đơn vị tính và số lượng vật tư không âm")
	}
	id, err := bson.ObjectIDFromHex(profileID)
	if err != nil {
		return mongo.ErrNoDocuments
	}
	for attempt := 0; attempt < 5; attempt++ {
		var profile types.MaterialsProfile
		if err := c.db.Collection("materials_profiles").FindOne(ctx, bson.M{"_id": id}).Decode(&profile); err != nil {
			return err
		}
		if request.Category == "consumable_supplies" {
			if profile.Estimate.ConsumableSupplies == nil {
				profile.Estimate.ConsumableSupplies = map[string]types.Material{}
			}
			profile.Estimate.ConsumableSupplies[request.Material.Name] = request.Material
		} else {
			if profile.Estimate.ReplacementMaterials == nil {
				profile.Estimate.ReplacementMaterials = map[string]types.Material{}
			}
			profile.Estimate.ReplacementMaterials[request.Material.Name] = request.Material
		}
		filter := bson.M{"_id": id, "version": profile.Version}
		if profile.Version == 0 {
			filter = bson.M{"_id": id, "$or": bson.A{bson.M{"version": 0}, bson.M{"version": bson.M{"$exists": false}}}}
		}
		result, err := c.db.Collection("materials_profiles").UpdateOne(ctx, filter, bson.M{
			"$set": bson.M{"estimate": profile.Estimate}, "$inc": bson.M{"version": 1},
		})
		if err != nil {
			return err
		}
		if result.MatchedCount == 1 {
			return nil
		}
	}
	return ErrConcurrentUpdate
}
