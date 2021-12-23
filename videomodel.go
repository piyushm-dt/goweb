package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type VideoMetaData struct {
	ID primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Key string `json:"videotokenid,omitempty" bson:"videotokenid,omitempty"`
	Title string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Genre string `json:"genre" bson:"genre"`
}

