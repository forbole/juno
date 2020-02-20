package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Post struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id"`
	PostID            uint64             `json:"post_id" bson:"post_id"`                       // Unique id
	ParentID          uint64             `json:"parent_id" bson:"parent_id"`                   // Post of which this one is a comment
	Message           string             `json:"message" bson:"message"`                       // Message contained inside the post
	Created           time.Time          `json:"created" bson:"created"`                       // Block height at which the post has been created
	LastEdited        uint64             `json:"last_edited" bson:"last_edited"`               // Block height at which the post has been edited the last time
	AllowsComments    bool               `json:"allows_comments" bson:"allows_commets"`        // Tells if users can reference this PostID as the parent
	ExternalReference string             `json:"external_reference" bson:"external_reference"` // Used to know when to display this post
	Owner             string             `json:"owner" bson:"owner"`                           // Creator of the Post
}
