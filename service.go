package reactions

import (
	"github.com/shurcooL/users"
	"golang.org/x/net/context"
)

// Service defines methods of a reactions service.
type Service interface {
	// Get reactions for id at uri.
	// uri is clean '/'-separated URI. E.g., "example.com/page".
	Get(ctx context.Context, uri string, id string) ([]Reaction, error)

	// Toggle a reaction for id at uri.
	Toggle(ctx context.Context, uri string, id string, tr ToggleRequest) ([]Reaction, error)
}

// Reaction represents a single reaction, backed by 1 or more users.
type Reaction struct {
	Reaction EmojiID
	Users    []users.User // Length is 1 or more.
}

// TODO, THINK: Maybe keep the colons, i.e., ":+1:".
// EmojiID is the id of a reaction. For example, "+1".
type EmojiID string

// ToggleRequest is a request to toggle a reaction.
type ToggleRequest struct {
	Reaction EmojiID
}

// Validate returns non-nil error if the request is invalid.
func (ToggleRequest) Validate() error {
	// TODO: Maybe validate that the emojiID is one of supported ones.
	//       Or maybe not (unsupported ones can be handled by frontend component).
	//       That way custom emoji can be added/removed, etc. Figure out what the best thing to do is and do it.
	return nil
}
