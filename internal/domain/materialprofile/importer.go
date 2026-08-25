package materialprofile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/remiehneppo/material-management/types"
	"github.com/remiehneppo/material-management/utils"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrEquipmentMismatch = errors.New("mã phân cấp này đã thuộc về một thiết bị khác")

type Importer struct {
	client *mongo.Client
	db     *mongo.Database
}

type importProfile struct {
	Index         int64
	EquipmentName string
	Estimate      types.MaterialsForEquipment
}

func NewImporter(client *mongo.Client, db *mongo.Database) *Importer {
	return &Importer{client: client, db: db}
}

func (i *Importer) Import(ctx context.Context, maintenanceID, sector, sheetName string, source io.Reader) error {
	if !utils.Contains(types.SECTOR_LIST, sector) {
		return types.ErrInvalidSector
	}
	workbook, err := excelize.OpenReader(source)
	if err != nil {
		return err
	}
	defer workbook.Close()
	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return err
	}
	profiles, err := parseRows(rows)
	if err != nil {
		return err
	}
	session, err := i.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx context.Context) (interface{}, error) {
		maintenanceObjectID, err := bson.ObjectIDFromHex(maintenanceID)
		if err != nil || i.db.Collection("maintenances").FindOne(tx, bson.M{"_id": maintenanceObjectID}).Err() != nil {
			return nil, types.ErrMaintenanceNotFound
		}
		for _, incoming := range profiles {
			var equipment types.EquipmentMachinery
			err := i.db.Collection("equipment_machineries").FindOne(tx, bson.M{"name": incoming.EquipmentName, "sector": sector}).Decode(&equipment)
			if errors.Is(err, mongo.ErrNoDocuments) {
				result, insertErr := i.db.Collection("equipment_machineries").InsertOne(tx, bson.M{"name": incoming.EquipmentName, "sector": sector})
				if insertErr != nil {
					return nil, insertErr
				}
				equipment.ID = result.InsertedID.(bson.ObjectID).Hex()
			} else if err != nil {
				return nil, err
			}

			key := bson.M{"maintenance_instance_id": maintenanceID, "sector": sector, "index": incoming.Index}
			var existing types.MaterialsProfile
			err = i.db.Collection("materials_profiles").FindOne(tx, key).Decode(&existing)
			if errors.Is(err, mongo.ErrNoDocuments) {
				incomingProfile := types.MaterialsProfile{MaintenanceInstanceID: maintenanceID, EquipmentMachineryID: equipment.ID, Sector: sector, Index: incoming.Index, Estimate: incoming.Estimate,
					Reality: types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{}, ReplacementMaterials: map[string]types.Material{}}}
				if _, err := i.db.Collection("materials_profiles").InsertOne(tx, incomingProfile); err != nil {
					return nil, err
				}
				continue
			}
			if err != nil {
				return nil, err
			}
			if existing.EquipmentMachineryID != equipment.ID {
				return nil, ErrEquipmentMismatch
			}
			existing.Estimate = mergeEstimate(existing.Estimate, incoming.Estimate)
			id, _ := bson.ObjectIDFromHex(existing.ID)
			if _, err := i.db.Collection("materials_profiles").UpdateOne(tx, bson.M{"_id": id}, bson.M{"$set": bson.M{"estimate": existing.Estimate}}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func parseRows(rows [][]string) ([]importProfile, error) {
	indexPattern := regexp.MustCompile(`^\d+(?:\.\d+)*$`)
	profiles := []importProfile{}
	seen := map[int64]struct{}{}
	current := -1
	category := ""
	for rowNumber, row := range rows {
		if len(row) < 2 {
			continue
		}
		indexCell := strings.Trim(strings.TrimSpace(row[0]), ".")
		title := strings.TrimSpace(row[1])
		if indexPattern.MatchString(indexCell) {
			index, err := utils.StringToIndexPath(indexCell)
			if err != nil {
				return nil, fmt.Errorf("dòng %d: %w", rowNumber+1, err)
			}
			if _, exists := seen[index]; exists {
				return nil, fmt.Errorf("dòng %d: mã phân cấp %s bị trùng", rowNumber+1, indexCell)
			}
			seen[index] = struct{}{}
			profiles = append(profiles, importProfile{Index: index, EquipmentName: title, Estimate: types.MaterialsForEquipment{ConsumableSupplies: map[string]types.Material{}, ReplacementMaterials: map[string]types.Material{}}})
			current = len(profiles) - 1
			category = ""
			continue
		}
		lowerTitle := strings.ToLower(title)
		if strings.Contains(lowerTitle, types.LABEL_REPLACEMENT) {
			category = "replacement"
			continue
		}
		if strings.Contains(lowerTitle, types.LABEL_CONSUMABLE) {
			category = "consumable"
			continue
		}
		if current < 0 || indexCell != "-" || category == "" {
			continue
		}
		if len(row) < 4 {
			return nil, fmt.Errorf("dòng %d: vật tư phải có đơn vị tính và số lượng", rowNumber+1)
		}
		quantity, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
		if err != nil || quantity < 0 {
			return nil, fmt.Errorf("dòng %d: số lượng không hợp lệ", rowNumber+1)
		}
		material := types.Material{Name: title, Unit: strings.ToLower(strings.TrimSpace(row[2])), Quantity: quantity}
		if material.Name == "" || material.Unit == "" {
			return nil, fmt.Errorf("dòng %d: vật tư phải có tên và đơn vị tính", rowNumber+1)
		}
		if category == "consumable" {
			profiles[current].Estimate.ConsumableSupplies[title] = material
		} else {
			profiles[current].Estimate.ReplacementMaterials[title] = material
		}
	}
	if len(profiles) == 0 {
		return nil, errors.New("trang tính không có hồ sơ vật tư nào")
	}
	return profiles, nil
}

func mergeEstimate(current, patch types.MaterialsForEquipment) types.MaterialsForEquipment {
	if current.ConsumableSupplies == nil {
		current.ConsumableSupplies = map[string]types.Material{}
	}
	if current.ReplacementMaterials == nil {
		current.ReplacementMaterials = map[string]types.Material{}
	}
	for key, material := range patch.ConsumableSupplies {
		current.ConsumableSupplies[key] = material
	}
	for key, material := range patch.ReplacementMaterials {
		current.ReplacementMaterials[key] = material
	}
	return current
}
