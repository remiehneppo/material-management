package migration

import (
	"context"
	"errors"
	"fmt"

	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Conflict struct {
	Kind   string      `json:"kind"`
	Key    interface{} `json:"key"`
	Count  int         `json:"count,omitempty"`
	Detail string      `json:"detail,omitempty"`
}

type Report struct {
	ReplicaSet string     `json:"replica_set"`
	Conflicts  []Conflict `json:"conflicts"`
}

func (r Report) Ready() bool { return r.ReplicaSet != "" && len(r.Conflicts) == 0 }

type Runner struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewRunner(client *mongo.Client, db *mongo.Database) *Runner {
	return &Runner{client: client, db: db}
}

func (r *Runner) Preflight(ctx context.Context) (Report, error) {
	report := Report{Conflicts: []Conflict{}}
	var hello struct {
		SetName string `bson:"setName"`
	}
	if err := r.db.RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&hello); err != nil {
		return report, err
	}
	report.ReplicaSet = hello.SetName
	if hello.SetName == "" {
		report.Conflicts = append(report.Conflicts, Conflict{Kind: "replica_set", Detail: "MongoDB is not a replica set"})
	}

	checks := []struct {
		collection, kind string
		match            bson.M
		group            bson.M
	}{
		{"users", "duplicate_username", bson.M{"username": bson.M{"$type": "string"}}, bson.M{"_id": "$username", "count": bson.M{"$sum": 1}}},
		{"materials_profiles", "duplicate_index_path", bson.M{}, bson.M{"_id": bson.M{"maintenance_instance_id": "$maintenance_instance_id", "sector": "$sector", "index": "$index"}, "count": bson.M{"$sum": 1}}},
		{"materials_requests", "duplicate_request_number", bson.M{"num_of_request": bson.M{"$gt": 0}}, bson.M{"_id": bson.M{"maintenance_instance_id": "$maintenance_instance_id", "num_of_request": "$num_of_request"}, "count": bson.M{"$sum": 1}}},
	}
	for _, check := range checks {
		cursor, err := r.db.Collection(check.collection).Aggregate(ctx, mongo.Pipeline{{{Key: "$match", Value: check.match}}, {{Key: "$group", Value: check.group}}, {{Key: "$match", Value: bson.M{"count": bson.M{"$gt": 1}}}}})
		if err != nil {
			return report, err
		}
		var rows []struct {
			ID    interface{} `bson:"_id"`
			Count int         `bson:"count"`
		}
		if err := cursor.All(ctx, &rows); err != nil {
			return report, err
		}
		for _, row := range rows {
			report.Conflicts = append(report.Conflicts, Conflict{Kind: check.kind, Key: row.ID, Count: row.Count})
		}
	}

	profiles, err := r.db.Collection("materials_profiles").Find(ctx, bson.M{})
	if err != nil {
		return report, err
	}
	var profileRows []types.MaterialsProfile
	if err := profiles.All(ctx, &profileRows); err != nil {
		return report, err
	}
	for _, profile := range profileRows {
		equipmentID, parseErr := bson.ObjectIDFromHex(profile.EquipmentMachineryID)
		if parseErr != nil || r.db.Collection("equipment_machineries").FindOne(ctx, bson.M{"_id": equipmentID}).Err() != nil {
			report.Conflicts = append(report.Conflicts, Conflict{Kind: "unmapped_equipment_machinery", Key: profile.ID})
		}
	}

	requests, err := r.db.Collection("materials_requests").Find(ctx, bson.M{})
	if err != nil {
		return report, err
	}
	var requestRows []types.MaterialRequest
	if err := requests.All(ctx, &requestRows); err != nil {
		return report, err
	}
	for _, request := range requestRows {
		if request.RequesterUserID == "" && r.db.Collection("users").FindOne(ctx, bson.M{"username": request.RequestedBy}).Err() != nil {
			report.Conflicts = append(report.Conflicts, Conflict{Kind: "unmapped_requester", Key: request.ID, Detail: request.RequestedBy})
		}
		for profileID := range request.MaterialsForEquipment {
			id, parseErr := bson.ObjectIDFromHex(profileID)
			if parseErr != nil || r.db.Collection("materials_profiles").FindOne(ctx, bson.M{"_id": id}).Err() != nil {
				report.Conflicts = append(report.Conflicts, Conflict{Kind: "unmapped_request_profile", Key: request.ID, Detail: profileID})
			}
		}
	}
	return report, nil
}

func (r *Runner) Apply(ctx context.Context) error {
	report, err := r.Preflight(ctx)
	if err != nil {
		return err
	}
	if !report.Ready() {
		return fmt.Errorf("migration preflight failed with %d conflict(s)", len(report.Conflicts))
	}
	if _, err := r.db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetName("unique_username").SetUnique(true).
			SetPartialFilterExpression(bson.M{"username": bson.M{"$type": "string"}}),
	}); err != nil {
		return err
	}
	session, err := r.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx context.Context) (interface{}, error) {
		cursor, err := r.db.Collection("materials_requests").Find(tx, bson.M{})
		if err != nil {
			return nil, err
		}
		var requests []types.MaterialRequest
		if err := cursor.All(tx, &requests); err != nil {
			return nil, err
		}
		maxByMaintenance := map[string]int{}
		for _, request := range requests {
			set := bson.M{}
			if request.RequesterUserID == "" {
				var user types.User
				if err := r.db.Collection("users").FindOne(tx, bson.M{"username": request.RequestedBy}).Decode(&user); err != nil {
					return nil, err
				}
				set["requester_user_id"] = user.ID
			}
			if request.Status == "" {
				if request.NumOfRequest > 0 {
					set["status"] = types.MATERIAL_REQUEST_ISSUED
					set["issued_at"] = request.RequestedAt
				} else {
					set["status"] = types.MATERIAL_REQUEST_DRAFT
				}
			}
			if len(set) > 0 {
				id, _ := bson.ObjectIDFromHex(request.ID)
				if _, err := r.db.Collection("materials_requests").UpdateOne(tx, bson.M{"_id": id}, bson.M{"$set": set}); err != nil {
					return nil, err
				}
			}
			if request.NumOfRequest > maxByMaintenance[request.MaintenanceInstanceID] {
				maxByMaintenance[request.MaintenanceInstanceID] = request.NumOfRequest
			}
		}
		for maintenanceID, last := range maxByMaintenance {
			if _, err := r.db.Collection("material_request_counters").UpdateOne(tx, bson.M{"_id": maintenanceID}, bson.M{"$max": bson.M{"last_number": last}}, options.UpdateOne().SetUpsert(true)); err != nil {
				return nil, err
			}
		}
		if _, err := r.db.Collection("user_sessions").DeleteMany(tx, bson.M{}); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	if _, err := r.db.Collection("materials_profiles").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "maintenance_instance_id", Value: 1}, {Key: "sector", Value: 1}, {Key: "index", Value: 1}},
		Options: options.Index().SetName("material_profile_business_key").SetUnique(true),
	}); err != nil {
		return err
	}
	_, err = r.db.Collection("materials_requests").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "maintenance_instance_id", Value: 1}, {Key: "num_of_request", Value: 1}},
		Options: options.Index().SetName("issued_request_number").SetUnique(true).SetPartialFilterExpression(bson.M{"status": types.MATERIAL_REQUEST_ISSUED}),
	})
	return err
}
