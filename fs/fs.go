// Package fs implements reactions.Service using a virtual filesystem.
package fs

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/shurcooL/reactions"
	"github.com/shurcooL/users"
	"github.com/shurcooL/webdavfs/vfsutil"
	"golang.org/x/net/webdav"
)

// NewService creates a virtual filesystem-backed reactions.Service using root for storage.
func NewService(root webdav.FileSystem, users users.Service) (reactions.Service, error) {
	return service{
		fs:    root,
		users: users,
	}, nil
}

type service struct {
	fs webdav.FileSystem

	users users.Service
}

func (s service) List(ctx context.Context, uri string) (map[string][]reactions.Reaction, error) {
	rm := make(map[string][]reactions.Reaction)
	fis, err := vfsutil.ReadDir(ctx, s.fs, reactablePath(uri))
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		// Get from storage.
		var reactable reactable
		err := jsonDecodeFile(ctx, s.fs, path.Join(reactablePath(uri), fi.Name()), &reactable)
		if err != nil {
			return nil, err
		}

		var rs []reactions.Reaction
		for _, r := range reactable.Reactions {
			reaction := reactions.Reaction{
				Reaction: r.EmojiID,
			}
			for _, u := range r.Authors {
				reactionAuthor := u.UserSpec()
				// TODO: Since we're potentially getting many of the same users multiple times here, consider caching them locally.
				reaction.Users = append(reaction.Users, s.user(ctx, reactionAuthor))
			}
			rs = append(rs, reaction)
		}
		rm[reactable.ID] = rs
	}

	return rm, nil
}

func (s service) Get(ctx context.Context, uri string, id string) ([]reactions.Reaction, error) {
	// Get from storage.
	var reactable reactable
	err := jsonDecodeFileNotDir(ctx, s.fs, path.Join(reactablePath(uri), sanitize(id)), &reactable)
	if err == errIsDir {
		return nil, os.ErrNotExist
	} else if err != nil {
		return nil, err
	}
	if reactable.ID != id {
		return nil, os.ErrNotExist
	}

	var rs []reactions.Reaction
	for _, r := range reactable.Reactions {
		reaction := reactions.Reaction{
			Reaction: r.EmojiID,
		}
		for _, u := range r.Authors {
			reactionAuthor := u.UserSpec()
			// TODO: Since we're potentially getting many of the same users multiple times here, consider caching them locally.
			reaction.Users = append(reaction.Users, s.user(ctx, reactionAuthor))
		}
		rs = append(rs, reaction)
	}
	return rs, nil
}

// canReact returns nil error if authenticatedUser is authorized to react to an entry.
// It returns os.ErrPermission or an error that happened in other cases.
func canReact(authenticatedUser users.UserSpec) error {
	if authenticatedUser.ID == 0 {
		// Not logged in, cannot react to anything.
		return os.ErrPermission
	}
	return nil
}

func (s service) Toggle(ctx context.Context, uri string, id string, tr reactions.ToggleRequest) ([]reactions.Reaction, error) {
	currentUser, err := s.users.GetAuthenticatedSpec(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser.ID == 0 {
		return nil, os.ErrPermission
	}

	err = tr.Validate()
	if err != nil {
		return nil, err
	}

	// Get from storage.
	var reactable reactable
	err = jsonDecodeFileNotDir(ctx, s.fs, path.Join(reactablePath(uri), sanitize(id)), &reactable)
	if err == errIsDir {
		return nil, os.ErrNotExist
	} else if err != nil {
		return nil, err
	}
	if reactable.ID != id {
		return nil, os.ErrNotExist
	}

	// Authorization check.
	if err := canReact(currentUser); err != nil {
		return nil, err
	}

	// Apply edits.
	err = toggleReaction(&reactable, currentUser, tr.Reaction)
	if err != nil {
		return nil, err
	}

	// Commit to storage.
	err = jsonEncodeFile(ctx, s.fs, path.Join(reactablePath(uri), sanitize(id)), reactable)
	if err != nil {
		return nil, err
	}

	var rs []reactions.Reaction
	for _, r := range reactable.Reactions {
		reaction := reactions.Reaction{
			Reaction: r.EmojiID,
		}
		for _, u := range r.Authors {
			reactionAuthor := u.UserSpec()
			// TODO: Since we're potentially getting many of the same users multiple times here, consider caching them locally.
			reaction.Users = append(reaction.Users, s.user(ctx, reactionAuthor))
		}
		rs = append(rs, reaction)
	}
	return rs, nil
}

// toggleReaction toggles reaction emojiID for specified user u.
// If user is creating a new reaction, they get added to the end of reaction authors.
func toggleReaction(c *reactable, u users.UserSpec, emojiID reactions.EmojiID) error {
	reactionsFromUser := 0
reactionsLoop:
	for _, r := range c.Reactions {
		for _, author := range r.Authors {
			if author.Equal(u) {
				reactionsFromUser++
				continue reactionsLoop
			}
		}
	}

	for i := range c.Reactions {
		if c.Reactions[i].EmojiID == emojiID {
			// Toggle this user's reaction.
			switch reacted := contains(c.Reactions[i].Authors, u); {
			case reacted == -1:
				// Add this reaction.
				if reactionsFromUser >= 20 {
					// TODO: Propagate this error as 400 Bad Request to frontend.
					return errors.New("too many reactions from same user")
				}
				c.Reactions[i].Authors = append(c.Reactions[i].Authors, fromUserSpec(u))
			default:
				// Remove this reaction. Delete without preserving order.
				c.Reactions[i].Authors[reacted] = c.Reactions[i].Authors[len(c.Reactions[i].Authors)-1]
				c.Reactions[i].Authors = c.Reactions[i].Authors[:len(c.Reactions[i].Authors)-1]

				// If there are no more authors backing it, this reaction goes away.
				if len(c.Reactions[i].Authors) == 0 {
					c.Reactions, c.Reactions[len(c.Reactions)-1] = append(c.Reactions[:i], c.Reactions[i+1:]...), reaction{} // Delete preserving order.
				}
			}
			return nil
		}
	}

	// If we get here, this is the first reaction of its kind.
	// Add it to the end of the list.
	if reactionsFromUser >= 20 {
		// TODO: Propagate this error as 400 Bad Request to frontend.
		return errors.New("too many reactions from same user")
	}
	c.Reactions = append(c.Reactions,
		reaction{
			EmojiID: emojiID,
			Authors: []userSpec{fromUserSpec(u)},
		},
	)
	return nil
}

// contains returns index of e in set, or -1 if it's not there.
func contains(set []userSpec, e users.UserSpec) int {
	for i, v := range set {
		if v.Equal(e) {
			return i
		}
	}
	return -1
}

/*func (s service) createNamespace(uri string) error {
	// Only needed for first issue in the repo.
	// TODO: Can this be better?
	return os.MkdirAll(filepath.Join(s.root, filepath.FromSlash(uri), issuesDir), 0755)
}*/
