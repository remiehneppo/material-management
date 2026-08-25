package materialrequest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	requestsCollection = "materials_requests"
	profilesCollection = "materials_profiles"
	usersCollection    = "users"
	countersCollection = "material_request_counters"
)

var (
	ErrDraftRequired     = errors.New("chỉ có thể ban hành yêu cầu vật tư đang ở trạng thái chờ")
	ErrRequesterMismatch = errors.New("chỉ người lập yêu cầu mới có thể ban hành yêu cầu vật tư này")
)

// Issuer owns the atomic Draft -> Issued transition. Its Interface is one
// operation because numbering, Reality accumulation and lifecycle state must
// never be invoked independently.
type Issuer struct {
	client *mongo.Client
	db     *mongo.Database
	now    func() time.Time
}

func NewIssuer(client *mongo.Client, db *mongo.Database) *Issuer {
	return &Issuer{client: client, db: db, now: time.Now}
}

func (i *Issuer) EnsureIndexes(ctx context.Context) error {
	_, err := i.db.Collection(requestsCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "maintenance_instance_id", Value: 1}, {Key: "num_of_request", Value: 1}},
		Options: options.Index().SetName("issued_request_number").SetUnique(true).
			SetPartialFilterExpression(bson.M{"status": types.MATERIAL_REQUEST_ISSUED}),
	})
	return err
}

func (i *Issuer) Issue(ctx context.Context, requestID, requesterUserID string) (*types.IssueMaterialRequestResponse, error) {
	requestObjectID, err := bson.ObjectIDFromHex(requestID)
	if err != nil {
		return nil, types.ErrMaterialRequestNotFound
	}

	session, err := i.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(tx context.Context) (interface{}, error) {
		var request types.MaterialRequest
		if err := i.db.Collection(requestsCollection).FindOne(tx, bson.M{"_id": requestObjectID}).Decode(&request); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, types.ErrMaterialRequestNotFound
			}
			return nil, err
		}
		if request.RequesterUserID == "" || request.RequesterUserID != requesterUserID {
			return nil, ErrRequesterMismatch
		}
		if request.Status == types.MATERIAL_REQUEST_ISSUED || request.NumOfRequest > 0 {
			return issueResponse(&request), nil
		}
		if request.Status != types.MATERIAL_REQUEST_DRAFT {
			return nil, ErrDraftRequired
		}
		userObjectID, err := bson.ObjectIDFromHex(requesterUserID)
		if err != nil {
			return nil, types.ErrUnauthorized
		}
		if err := i.db.Collection(usersCollection).FindOne(tx, bson.M{"_id": userObjectID}).Err(); err != nil {
			return nil, types.ErrUnauthorized
		}

		var counter struct {
			Last int `bson:"last_number"`
		}
		err = i.db.Collection(countersCollection).FindOneAndUpdate(tx,
			bson.M{"_id": request.MaintenanceInstanceID},
			bson.M{"$inc": bson.M{"last_number": 1}},
			options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
		).Decode(&counter)
		if err != nil {
			return nil, err
		}

		for profileID, requested := range request.MaterialsForEquipment {
			profileObjectID, err := bson.ObjectIDFromHex(profileID)
			if err != nil {
				return nil, types.ErrSomeMaterialsProfileNotFound
			}
			var profile types.MaterialsProfile
			if err := i.db.Collection(profilesCollection).FindOne(tx, bson.M{
				"_id": profileObjectID, "maintenance_instance_id": request.MaintenanceInstanceID, "sector": request.Sector,
			}).Decode(&profile); err != nil {
				return nil, types.ErrSomeMaterialsProfileNotFound
			}
			if !MaterialsBelongToProfile(&profile, requested) {
				return nil, types.ErrMaterialNotInProfile
			}
			profile.Reality = Accumulate(profile.Reality, requested)
			if _, err := i.db.Collection(profilesCollection).UpdateOne(tx, bson.M{"_id": profileObjectID}, bson.M{"$set": bson.M{"reality": profile.Reality}}); err != nil {
				return nil, err
			}
		}

		issuedAt := i.now().Unix()
		update := bson.M{"$set": bson.M{
			"num_of_request": counter.Last,
			"status":         types.MATERIAL_REQUEST_ISSUED,
			"issued_at":      issuedAt,
			"issued_by":      requesterUserID,
		}}
		updated, err := i.db.Collection(requestsCollection).UpdateOne(tx, bson.M{
			"_id": requestObjectID, "status": types.MATERIAL_REQUEST_DRAFT, "num_of_request": 0,
		}, update)
		if err != nil {
			return nil, err
		}
		if updated.ModifiedCount != 1 {
			return nil, fmt.Errorf("%w: concurrent transition", ErrDraftRequired)
		}
		request.NumOfRequest = counter.Last
		request.Status = types.MATERIAL_REQUEST_ISSUED
		request.IssuedAt = issuedAt
		request.IssuedBy = requesterUserID
		return issueResponse(&request), nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.IssueMaterialRequestResponse), nil
}

func issueResponse(request *types.MaterialRequest) *types.IssueMaterialRequestResponse {
	return &types.IssueMaterialRequestResponse{
		RequestNumber: request.NumOfRequest,
		Status:        request.Status,
		IssuedAt:      request.IssuedAt,
		IssuedBy:      request.IssuedBy,
	}
}

func Accumulate(current, delta types.MaterialsForEquipment) types.MaterialsForEquipment {
	if current.ConsumableSupplies == nil {
		current.ConsumableSupplies = map[string]types.Material{}
	}
	if current.ReplacementMaterials == nil {
		current.ReplacementMaterials = map[string]types.Material{}
	}
	mergeMaterials(current.ConsumableSupplies, delta.ConsumableSupplies)
	mergeMaterials(current.ReplacementMaterials, delta.ReplacementMaterials)
	return current
}

func mergeMaterials(target, delta map[string]types.Material) {
	for key, material := range delta {
		existing := target[key]
		if existing.Name == "" {
			existing.Name = material.Name
			existing.Unit = material.Unit
		}
		existing.Quantity += material.Quantity
		target[key] = existing
	}
}
