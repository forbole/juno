package types

import (
	"time"

	"github.com/desmos-labs/desmos/x/posts"
)

// NewPostFromMsg allows to create a new Desmos post instance from the given msg.
// This is required as the postID is not present withing them message itself, but is instead
// returned to the client through the events or logs array.
func NewPostFromMsg(postID uint64, msg posts.MsgCreatePost) posts.Post {
	return posts.Post{
		PostID:         posts.PostID(postID),
		ParentID:       msg.ParentID,
		Message:        msg.Message,
		Created:        msg.CreationDate,
		LastEdited:     time.Time{},
		AllowsComments: msg.AllowsComments,
		Subspace:       msg.Subspace,
		OptionalData:   msg.OptionalData,
		Creator:        msg.Creator,
	}
}
