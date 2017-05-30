package fs

import (
	"path"
	"strings"
	"unicode"

	"github.com/shurcooL/reactions"
	"github.com/shurcooL/users"
)

// userSpec is an on-disk representation of users.UserSpec.
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

// reactable is an on-disk representation of []reactions.Reaction.
type reactable struct {
	ID        string
	Reactions []reaction `json:",omitempty"`
}

// reaction is an on-disk representation of reactions.Reaction.
type reaction struct {
	EmojiID reactions.EmojiID
	Authors []userSpec // First entry is first person who reacted.
}

// Tree layout:
//
// 	root
// 	└── domain.com
// 	    └── path
// 	        ├── a - encoded reactable
// 	        ├── b
// 	        └── c

// TODO.
func reactablePath(uri string) string {
	var elems []string
	for _, e := range strings.Split(uri, "/") {
		elems = append(elems, sanitize(e))
	}
	return path.Join(elems...)
}

// TODO.
func sanitize(text string) string {
	var anchorName []rune
	var futureDash = false
	for _, r := range text {
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
