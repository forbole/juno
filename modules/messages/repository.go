package messages

import "github.com/forbole/juno/v5/types"

type MessageRepository interface {
	// SaveMessage saves the given message inside the database
	SaveMessage(msg *types.Message) error
}
