package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"healthcare-system-sawtooth/client/db"
)

const (
	Unset    int = 0
	Accepted int = 1
	Rejected int = 2
)

// Request model for MongoDB
type Request struct {
	OID          *primitive.ObjectID `json:"OID" bson:"_id,omitempty"`
	Hash         string              `json:"hash" bson:"hash,omitempty"`
	Name         string              `json:"name" bson:"name,omitempty"`
	RequestFrom  string              `json:"request_from" bson:"request_from,omitempty"`
	UsernameFrom string              `json:"username_from" bson:"username_from,omitempty"`
	UsernameTo   string              `json:"username_to" bson:"username_to,omitempty"`
	Status       int                 `json:"status" bson:"status,omitempty"`
	AccessType   int                 `json:"access_type" bson:"access_type,omitempty"`
}

func UpsertRequests(ctx context.Context, pms []*Request) (int64, error) {
	col, err := getMongoRequestCollection(ctx)
	if err != nil {
		return 0, err
	}

	var operations []mongo.WriteModel
	for _, pm := range pms {
		fmt.Printf("%+v", pm)
		op := mongo.NewUpdateOneModel()

		if pm.OID != nil {
			op.SetFilter(bson.M{"_id": pm.OID})
		}

		update := bson.M{}
		if pm.Hash != "" {
			update["hash"] = pm.Hash
		}
		if pm.RequestFrom != "" {
			update["request_from"] = pm.RequestFrom
		}
		if pm.UsernameFrom != "" {
			update["username_from"] = pm.UsernameFrom
		}
		if pm.UsernameTo != "" {
			update["username_to"] = pm.UsernameTo
		}
		if pm.Status != 0 {
			update["status"] = pm.Status
		}
		if pm.AccessType != 0 {
			update["access_type"] = pm.AccessType
		}
		if pm.AccessType != 0 {
			update["name"] = pm.Name
		}
		op.SetUpdate(bson.M{"$set": update})
		op.SetUpsert(true)
		operations = append(operations, op)
	}

	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(true)

	res, err := col.BulkWrite(ctx, operations, &bulkOption)
	if err != nil {
		return 0, err
	}
	return res.UpsertedCount, nil
}

// Save stores Request into the database
func (d *Request) Save() (*primitive.ObjectID, error) {

	ctx, cancel := db.GetMongoContext()
	defer cancel()
	col, err := getMongoRequestCollection(ctx)
	if err != nil {
		return nil, err
	}

	res, err := col.InsertOne(ctx, d)
	if err != nil {
		return nil, err
	}

	objID := res.InsertedID.(primitive.ObjectID)
	d.OID = &objID
	return &objID, nil
}

// GetRequestsByHashes gets data from the database
func GetRequestsByRequestFrom(ctx context.Context, requestFrom []string) ([]*Request, error) {
	col, err := getMongoRequestCollection(ctx)
	if err != nil {
		return nil, err
	}
	var pms []*Request
	c, err := col.Find(ctx, bson.M{"request_from": bson.M{"$in": requestFrom}})
	if err != nil {
		return nil, err
	}
	err = c.All(ctx, &pms)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return pms, nil
}

func GetRequestByOID(ctx context.Context, orgOID *primitive.ObjectID) (*Request, error) {
	ctx, cancel := db.GetMongoContext()
	defer cancel()
	col, err := getMongoRequestCollection(ctx)
	if err != nil {
		return nil, err
	}

	var app Request
	if err := db.GetMongoObject(ctx, col, orgOID).Decode(&app); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &app, nil
}

// Get table name
func getMongoRequestCollection(ctx context.Context) (*mongo.Collection, error) {
	return db.GetMongoCollection(ctx, db.MongoRequestCollection)
}
