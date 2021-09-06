package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"healthcare-system-sawtooth/client/db"
)

// Data model for MongoDB
type Data struct {
	OID        *primitive.ObjectID `json:"OID" bson:"_id,omitempty"`
	Hash       string              `json:"hash"`
	Name       string              `json:"name"`
	Payload    string              `json:"payload"`
	Expiration int64               `json:"expiration"`
}

// Save stores data into the database
func (d *Data) Save() (*primitive.ObjectID, error) {

	ctx, cancel := db.GetMongoContext()
	defer cancel()
	col, err := getMongoDataCollection(ctx)
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

// GetDataByHashes gets data from the database
func GetDataByHashes(ctx context.Context, hashes []string) ([]*Data, error) {
	col, err := getMongoDataCollection(ctx)
	if err != nil {
		return nil, err
	}
	var pms []*Data
	c, err := col.Find(ctx, bson.M{"hash": bson.M{"$in": hashes}})
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

func GetDataByExpiration(ctx context.Context, now int64) ([]*Data, error) {
	col, err := getMongoDataCollection(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"$and": []bson.M{{"expiration": bson.M{"$ne": 0}}, {"expiration": bson.M{"$lte": now}}}}

	var pms []*Data
	c, err := col.Find(ctx, filter)
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

func DeleteDatasByOid(oids []*primitive.ObjectID) error {
	ctx, cancel := db.GetMongoContext()
	defer cancel()
	col, err := getMongoDataCollection(ctx)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": bson.M{"$in": oids}}
	_, err = col.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

// Get table name
func getMongoDataCollection(ctx context.Context) (*mongo.Collection, error) {
	return db.GetMongoCollection(ctx, db.MongoDataCollection)
}
