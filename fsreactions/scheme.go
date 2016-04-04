package fsreactions

import (
	"path"
	"strings"
	"unicode"

	"github.com/shurcooL/reactions"
	"github.com/shurcooL/users"
)

// userSpec is an on-disk representation of a specification for a user.
type userSpec struct {
	ID     uint64
	Domain string `json:",omitempty"`
}

func fromUserSpec(us users.UserSpec) userSpec {
	return userSpec{ID: us.ID, Domain: us.Domain}
}

func (us userSpec) UserSpec() users.UserSpec {
	return users.UserSpec{ID: us.ID, Domain: us.Domain}
}

func (us userSpec) Equal(other users.UserSpec) bool {
	return us.Domain == other.Domain && us.ID == other.ID
}

// reactable is an on-disk representation of a reactable.
type reactable struct {
	Reactions []reaction `json:",omitempty"`
}

// reaction is an on-disk representation of a reaction.
type reaction struct {
	EmojiID reactions.EmojiID
	Authors []userSpec // Order does not matter; this would be better represented as a set like map[userSpec]struct{}, but we're using JSON and it doesn't support that.
}

// TODO.
func reactablePath(uri string) string {
	var elems []string
	for _, e := range strings.Split(uri, "/") {
		elems = append(elems, sanitize(e))
	}
	return path.Join(elems...)
}

func sanitize(text string) string {
	var anchorName []rune
	var futureDash = false
	for _, r := range []rune(text) {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r) || r == '.':
			if futureDash && len(anchorName) > 0 {
				anchorName = append(anchorName, '-')
			}
			futureDash = false
			anchorName = append(anchorName, unicode.ToLower(r))
		default:
			futureDash = true
		}
	}
	return string(anchorName)
}
