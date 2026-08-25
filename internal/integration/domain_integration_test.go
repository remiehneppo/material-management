package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/internal/domain/materialprofile"
	"github.com/remiehneppo/material-management/internal/domain/materialrequest"
	domainsession "github.com/remiehneppo/material-management/internal/domain/session"
	domainuser "github.com/remiehneppo/material-management/internal/domain/user"
	"github.com/remiehneppo/material-management/internal/migration"
	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
	"github.com/remiehneppo/material-management/utils"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func testDatabase(t *testing.T) (*mongo.Client, *mongo.Database) {
	t.Helper()
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		t.Skip("MONGO_INTEGRATION_URI is not set")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetBSONOptions(&options.BSONOptions{ObjectIDAsHexString: true}))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("material_management_integration_" + fmt.Sprint(time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()); _ = client.Disconnect(context.Background()) })
	return client, db
}

func TestBootstrapAdminCreatesOnceWithoutOverwriting(t *testing.T) {
	_, db := testDatabase(t)
	ctx := context.Background()
	initial := domainuser.BootstrapAdminConfig{
		Username:  "admin",
		Password:  "Admin@123456",
		FullName:  "Initial administrator",
		Workspace: "System",
	}
	created, err := domainuser.BootstrapAdmin(ctx, db, initial)
	if err != nil || !created {
		t.Fatalf("first bootstrap created=%v err=%v", created, err)
	}

	var first types.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"username": initial.Username}).Decode(&first); err != nil {
		t.Fatal(err)
	}
	valid, err := domainsession.VerifyPassword(first.Password, initial.Password)
	if err != nil || !valid {
		t.Fatalf("bootstrap password is not a valid Argon2id hash: valid=%v err=%v", valid, err)
	}
	if first.WorkspaceRole != types.USER_ROLE_ADMIN {
		t.Fatalf("workspace_role=%q", first.WorkspaceRole)
	}

	repeated := initial
	repeated.Password = "Changed@123456"
	repeated.FullName = "Changed administrator"
	created, err = domainuser.BootstrapAdmin(ctx, db, repeated)
	if err != nil || created {
		t.Fatalf("repeat bootstrap created=%v err=%v", created, err)
	}
	var unchanged types.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"username": initial.Username}).Decode(&unchanged); err != nil {
		t.Fatal(err)
	}
	if unchanged.Password != first.Password || unchanged.FullName != first.FullName {
		t.Fatal("repeat bootstrap overwrote the existing administrator")
	}
}

func TestBootstrapAdminCanBeDisabled(t *testing.T) {
	created, err := domainuser.BootstrapAdmin(context.Background(), nil, domainuser.BootstrapAdminConfig{})
	if err != nil || created {
		t.Fatalf("disabled bootstrap created=%v err=%v", created, err)
	}
}

func TestBootstrapAdminRejectsPartialCredentials(t *testing.T) {
	_, err := domainuser.BootstrapAdmin(context.Background(), nil, domainuser.BootstrapAdminConfig{Username: "admin"})
	if err == nil {
		t.Fatal("partial bootstrap credentials were accepted")
	}
}

func TestConcurrentIssuanceIsGaplessAtomicAndIdempotent(t *testing.T) {
	client, db := testDatabase(t)
	ctx := context.Background()
	userID, profileID := bson.NewObjectID(), bson.NewObjectID()
	_, _ = db.Collection("users").InsertOne(ctx, bson.M{"_id": userID, "username": "requester"})
	_, _ = db.Collection("materials_profiles").InsertOne(ctx, bson.M{"_id": profileID, "maintenance_instance_id": "maintenance", "equipment_machinery_id": "equipment", "sector": "Cơ khí", "estimate": bson.M{"consumable_supplies": bson.M{"oil": bson.M{"name": "oil", "unit": "l", "quantity": 10}}}, "reality": bson.M{}})
	requestIDs := []bson.ObjectID{bson.NewObjectID(), bson.NewObjectID()}
	for _, id := range requestIDs {
		_, err := db.Collection("materials_requests").InsertOne(ctx, bson.M{"_id": id, "maintenance_instance_id": "maintenance", "sector": "Cơ khí", "status": "draft", "num_of_request": 0, "requester_user_id": userID.Hex(), "materials_for_equipment": bson.M{profileID.Hex(): bson.M{"consumable_supplies": bson.M{"oil": bson.M{"name": "oil", "unit": "l", "quantity": 1}}}}})
		if err != nil {
			t.Fatal(err)
		}
	}
	issuer := materialrequest.NewIssuer(client, db)
	if err := issuer.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	results := make([]int, 2)
	errorsFound := make([]error, 2)
	var wait sync.WaitGroup
	for index := range requestIDs {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			result, err := issuer.Issue(ctx, requestIDs[index].Hex(), userID.Hex())
			errorsFound[index] = err
			if result != nil {
				results[index] = result.RequestNumber
			}
		}(index)
	}
	wait.Wait()
	for _, err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Ints(results)
	if results[0] != 1 || results[1] != 2 {
		t.Fatalf("numbers=%v", results)
	}
	retry, err := issuer.Issue(ctx, requestIDs[0].Hex(), userID.Hex())
	if err != nil || retry.RequestNumber == 0 {
		t.Fatalf("retry=%+v err=%v", retry, err)
	}
	var profile types.MaterialsProfile
	if err := db.Collection("materials_profiles").FindOne(ctx, bson.M{"_id": profileID}).Decode(&profile); err != nil {
		t.Fatal(err)
	}
	if profile.Reality.ConsumableSupplies["oil"].Quantity != 2 {
		t.Fatalf("Reality=%+v", profile.Reality)
	}

	badID := bson.NewObjectID()
	_, _ = db.Collection("materials_requests").InsertOne(ctx, bson.M{"_id": badID, "maintenance_instance_id": "maintenance", "sector": "Cơ khí", "status": "draft", "num_of_request": 0, "requester_user_id": userID.Hex(), "materials_for_equipment": bson.M{profileID.Hex(): bson.M{"consumable_supplies": bson.M{"oil": bson.M{"name": "oil", "unit": "l", "quantity": 9}}}, bson.NewObjectID().Hex(): bson.M{}}})
	if _, err := issuer.Issue(ctx, badID.Hex(), userID.Hex()); err == nil {
		t.Fatal("invalid profile did not roll back")
	}
	var counter struct {
		Last int `bson:"last_number"`
	}
	if err := db.Collection("material_request_counters").FindOne(ctx, bson.M{"_id": "maintenance"}).Decode(&counter); err != nil {
		t.Fatal(err)
	}
	if counter.Last != 2 {
		t.Fatalf("counter advanced after rollback: %d", counter.Last)
	}
}

type fakeAccessIssuer struct{}

func (fakeAccessIssuer) GenerateAccessTokenForSession(_ *types.User, sessionID string) (string, error) {
	return "access-" + sessionID, nil
}

func TestSessionRotationRejectsReplayWithoutRevokingSuccessor(t *testing.T) {
	_, db := testDatabase(t)
	ctx := context.Background()
	password, _ := domainsession.HashPassword("password1")
	userID := bson.NewObjectID()
	_, _ = db.Collection("users").InsertOne(ctx, bson.M{"_id": userID, "username": "tester", "password": password})
	manager := domainsession.NewManager(db, fakeAccessIssuer{}, time.Hour)
	login, err := manager.Login(ctx, "tester", "password1")
	if err != nil {
		t.Fatal(err)
	}
	sessionID := login.AccessToken[len("access-"):]
	if err := manager.ValidateAccessSession(ctx, sessionID, userID.Hex()); err != nil {
		t.Fatalf("new access session rejected: %v", err)
	}
	var wait sync.WaitGroup
	var tokens [2]*domainsession.Tokens
	var errs [2]error
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			tokens[index], errs[index] = manager.Refresh(ctx, login.RefreshToken)
		}(index)
	}
	wait.Wait()
	success := 0
	var successor string
	for index := range errs {
		if errs[index] == nil {
			success++
			successor = tokens[index].RefreshToken
		}
	}
	if success != 1 {
		t.Fatalf("success=%d errs=%v", success, errs)
	}
	if _, err := manager.Refresh(ctx, login.RefreshToken); err == nil {
		t.Fatal("replayed token accepted")
	}
	rotated, err := manager.Refresh(ctx, successor)
	if err != nil || rotated.RefreshToken == "" {
		t.Fatalf("successor was revoked: %v", err)
	}
	deviceA, err := manager.Login(ctx, "tester", "password1")
	if err != nil {
		t.Fatal(err)
	}
	deviceB, err := manager.Login(ctx, "tester", "password1")
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Logout(ctx, deviceA.RefreshToken); err != nil {
		t.Fatal(err)
	}
	deviceASessionID := deviceA.AccessToken[len("access-"):]
	if err := manager.ValidateAccessSession(ctx, deviceASessionID, userID.Hex()); err == nil {
		t.Fatal("logged-out access session remained active")
	}
	if _, err := manager.Refresh(ctx, deviceA.RefreshToken); err == nil {
		t.Fatal("logged-out device refreshed")
	}
	if _, err := manager.Refresh(ctx, deviceB.RefreshToken); err != nil {
		t.Fatalf("other device was revoked: %v", err)
	}

	legacyID := bson.NewObjectID()
	_, _ = db.Collection("users").InsertOne(ctx, bson.M{"_id": legacyID, "username": "legacyuser", "password": "password2"})
	if _, err := manager.Login(ctx, "legacyuser", "password2"); err != nil {
		t.Fatal(err)
	}
	var legacy types.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"_id": legacyID}).Decode(&legacy); err != nil {
		t.Fatal(err)
	}
	if len(legacy.Password) < 10 || legacy.Password[:10] != "$argon2id$" {
		t.Fatalf("plaintext password was not rehashed: %q", legacy.Password)
	}
}

func TestPasswordChangeRevokesAllUserSessions(t *testing.T) {
	client, db := testDatabase(t)
	ctx := context.Background()
	password, _ := domainsession.HashPassword("password1")
	userID, sessionID := bson.NewObjectID(), bson.NewObjectID()
	_, _ = db.Collection("users").InsertOne(ctx, bson.M{"_id": userID, "username": "tester", "password": password})
	_, _ = db.Collection("user_sessions").InsertOne(ctx, bson.M{
		"_id": sessionID, "user_id": userID.Hex(), "token_hash": "hash", "created_at": time.Now().Unix(), "expires_at": time.Now().Add(time.Hour).Unix(),
	})

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/user/change-password", bytes.NewBufferString(`{"old_password":"password1","new_password":"password2"}`))
	request.Header.Set("Content-Type", "application/json")
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = request
	ginContext.Set("user", &types.User{ID: userID.Hex()})
	domainuser.NewHandler(client, db).ChangePassword(ginContext)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if err := domainsession.NewManager(db, fakeAccessIssuer{}, time.Hour).ValidateAccessSession(ctx, sessionID.Hex(), userID.Hex()); err == nil {
		t.Fatal("password change left an existing session active")
	}
}

func TestEstimateImportMergesByIndexAndRollsBackMismatch(t *testing.T) {
	client, db := testDatabase(t)
	ctx := context.Background()
	maintenanceID, equipmentID := bson.NewObjectID(), bson.NewObjectID()
	_, _ = db.Collection("maintenances").InsertOne(ctx, bson.M{"_id": maintenanceID})
	_, _ = db.Collection("equipment_machineries").InsertOne(ctx, bson.M{"_id": equipmentID, "name": "Pump", "sector": "Cơ khí"})
	index, _ := utils.StringToIndexPath("1.1")
	_, _ = db.Collection("materials_profiles").InsertOne(ctx, bson.M{"maintenance_instance_id": maintenanceID.Hex(), "equipment_machinery_id": equipmentID.Hex(), "sector": "Cơ khí", "index": index, "estimate": bson.M{"consumable_supplies": bson.M{"old": bson.M{"name": "old", "unit": "l", "quantity": 1}}}, "reality": bson.M{}})
	importer := materialprofile.NewImporter(client, db)
	if err := importer.Import(ctx, maintenanceID.Hex(), "Cơ khí", "Estimate", workbook(t, [][]string{{"1.1", "Pump"}, {"", types.LABEL_CONSUMABLE}, {"-", "new", "l", "2"}, {"1.2", "Pump"}, {"", types.LABEL_REPLACEMENT}, {"-", "seal", "piece", "1"}})); err != nil {
		t.Fatal(err)
	}
	if count, _ := db.Collection("materials_profiles").CountDocuments(ctx, bson.M{}); count != 2 {
		t.Fatalf("profiles=%d", count)
	}
	var profile types.MaterialsProfile
	if err := db.Collection("materials_profiles").FindOne(ctx, bson.M{"index": index}).Decode(&profile); err != nil {
		t.Fatal(err)
	}
	if len(profile.Estimate.ConsumableSupplies) != 2 {
		t.Fatalf("estimate=%+v", profile.Estimate)
	}
	if err := importer.Import(ctx, maintenanceID.Hex(), "Cơ khí", "Estimate", workbook(t, [][]string{{"1.1", "Other"}, {"", types.LABEL_CONSUMABLE}, {"-", "bad", "l", "5"}})); err == nil {
		t.Fatal("equipment mismatch accepted")
	}
	if count, _ := db.Collection("equipment_machineries").CountDocuments(ctx, bson.M{"name": "Other"}); count != 0 {
		t.Fatalf("auto-created equipment survived rollback")
	}
}

func TestConcurrentCatalogUpsertsPreserveBothMaterials(t *testing.T) {
	_, db := testDatabase(t)
	ctx := context.Background()
	profileID := bson.NewObjectID()
	_, _ = db.Collection("materials_profiles").InsertOne(ctx, bson.M{
		"_id": profileID, "estimate": bson.M{"consumable_supplies": bson.M{}}, "reality": bson.M{},
	})
	catalog := materialprofile.NewCatalog(db)
	updates := []materialprofile.UpsertMaterialRequest{
		{Category: "consumable_supplies", Material: types.Material{Name: "grease", Unit: "kg", Quantity: 1}},
		{Category: "consumable_supplies", Material: types.Material{Name: "solvent", Unit: "l", Quantity: 2}},
	}
	errorsFound := make(chan error, len(updates))
	for _, update := range updates {
		go func(update materialprofile.UpsertMaterialRequest) {
			errorsFound <- catalog.UpsertEstimatedMaterial(ctx, profileID.Hex(), update)
		}(update)
	}
	for range updates {
		if err := <-errorsFound; err != nil {
			t.Fatal(err)
		}
	}
	var profile types.MaterialsProfile
	if err := db.Collection("materials_profiles").FindOne(ctx, bson.M{"_id": profileID}).Decode(&profile); err != nil {
		t.Fatal(err)
	}
	if _, ok := profile.Estimate.ConsumableSupplies["grease"]; !ok {
		t.Fatal("concurrent grease upsert was lost")
	}
	if _, ok := profile.Estimate.ConsumableSupplies["solvent"]; !ok {
		t.Fatal("concurrent solvent upsert was lost")
	}
}

func TestIssuedRequestRejectsDraftUpdateAndDelete(t *testing.T) {
	_, db := testDatabase(t)
	ctx := context.Background()
	wrapper := database.NewMongoDatabase(os.Getenv("MONGO_INTEGRATION_URI"), db.Name())
	if err := wrapper.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wrapper.Disconnect(context.Background()) })
	repo := repository.NewMaterialsRequestRepository(wrapper)
	requestID := bson.NewObjectID()
	ownerID := bson.NewObjectID().Hex()
	_, _ = db.Collection("materials_requests").InsertOne(ctx, bson.M{
		"_id": requestID, "status": types.MATERIAL_REQUEST_ISSUED, "num_of_request": 1, "requester_user_id": ownerID,
	})
	updated, err := repo.UpdateDraft(ctx, requestID.Hex(), ownerID, &types.MaterialRequestUpdate{Description: "stale update"})
	if err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Fatal("issued request accepted a draft update")
	}
	deleted, err := repo.DeleteDraft(ctx, requestID.Hex(), ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Fatal("issued request accepted a draft delete")
	}
}

func TestMigrationPreflightRejectsDuplicateUsernames(t *testing.T) {
	client, db := testDatabase(t)
	ctx := context.Background()
	_, _ = db.Collection("users").InsertMany(ctx, []interface{}{
		bson.M{"_id": bson.NewObjectID(), "username": "duplicate"},
		bson.M{"_id": bson.NewObjectID(), "username": "duplicate"},
	})
	report, err := migration.NewRunner(client, db).Preflight(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, conflict := range report.Conflicts {
		if conflict.Kind == "duplicate_username" {
			return
		}
	}
	t.Fatal("duplicate username was not reported")
}

func TestMigrationPreflightIsReadOnlyAndApplyIsIdempotent(t *testing.T) {
	client, db := testDatabase(t)
	ctx := context.Background()
	userID, requestID := bson.NewObjectID(), bson.NewObjectID()
	_, _ = db.Collection("users").InsertOne(ctx, bson.M{"_id": userID, "username": "legacy"})
	_, _ = db.Collection("materials_requests").InsertOne(ctx, bson.M{"_id": requestID, "maintenance_instance_id": "maintenance", "requested_by": "legacy", "requested_at": int64(123), "num_of_request": 4, "materials_for_equipment": bson.M{}})
	runner := migration.NewRunner(client, db)
	report, err := runner.Preflight(ctx)
	if err != nil || !report.Ready() {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	var before bson.M
	if err := db.Collection("materials_requests").FindOne(ctx, bson.M{"_id": requestID}).Decode(&before); err != nil {
		t.Fatal(err)
	}
	if _, exists := before["status"]; exists {
		t.Fatal("preflight wrote status")
	}
	if err := runner.Apply(ctx); err != nil {
		t.Fatal(err)
	}
	if err := runner.Apply(ctx); err != nil {
		t.Fatalf("second apply: %v", err)
	}
	var migrated types.MaterialRequest
	if err := db.Collection("materials_requests").FindOne(ctx, bson.M{"_id": requestID}).Decode(&migrated); err != nil {
		t.Fatal(err)
	}
	if migrated.Status != types.MATERIAL_REQUEST_ISSUED || migrated.RequesterUserID != userID.Hex() || migrated.IssuedAt != 123 {
		t.Fatalf("migrated=%+v", migrated)
	}
	var counter struct {
		Last int `bson:"last_number"`
	}
	if err := db.Collection("material_request_counters").FindOne(ctx, bson.M{"_id": "maintenance"}).Decode(&counter); err != nil || counter.Last != 4 {
		t.Fatalf("counter=%+v err=%v", counter, err)
	}
}

func workbook(t *testing.T, rows [][]string) *bytes.Reader {
	t.Helper()
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", "Estimate")
	for rowIndex, row := range rows {
		for columnIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+1)
			_ = file.SetCellValue("Estimate", cell, value)
		}
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buffer.Bytes())
}
