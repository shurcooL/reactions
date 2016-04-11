package fs

import (
	"reflect"
	"testing"

	"github.com/shurcooL/reactions"
	"github.com/shurcooL/users"
)

func TestToggleReaction(t *testing.T) {
	r := reactable{
		Reactions: []reaction{
			{EmojiID: reactions.EmojiID("bar"), Authors: []userSpec{{ID: 1}, {ID: 2}}},
			{EmojiID: reactions.EmojiID("baz"), Authors: []userSpec{{ID: 3}}},
		},
	}

	toggleReaction(&r, users.UserSpec{ID: 1}, reactions.EmojiID("foo"))
	toggleReaction(&r, users.UserSpec{ID: 1}, reactions.EmojiID("bar"))
	toggleReaction(&r, users.UserSpec{ID: 1}, reactions.EmojiID("baz"))
	toggleReaction(&r, users.UserSpec{ID: 2}, reactions.EmojiID("bar"))

	want := reactable{
		Reactions: []reaction{
			{EmojiID: reactions.EmojiID("baz"), Authors: []userSpec{{ID: 3}, {ID: 1}}},
			{EmojiID: reactions.EmojiID("foo"), Authors: []userSpec{{ID: 1}}},
		},
	}

	if got := r; !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot  %+v\nwant %+v", got.Reactions, want.Reactions)
	}
}
