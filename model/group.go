package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Group struct {
	Id          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name,omitempty" bson:"name,omitempty"`
	Description string               `json:"description,omitempty" bson:"description,omitempty"`
	Users       []primitive.ObjectID `json:"users,omitempty" bson:"users,omitempty"`
}
