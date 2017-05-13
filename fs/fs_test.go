package fs_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/shurcooL/reactions"
	"github.com/shurcooL/reactions/fs"
	"github.com/shurcooL/users"
	"github.com/shurcooL/webdavfs/vfsutil"
	"golang.org/x/net/webdav"
)

func TestService_notExist(t *testing.T) {
	memFS := webdav.NewMemFS()
	service, err := fs.NewService(memFS, mockUsers{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.List(context.Background(), "not/exist")
	if !os.IsNotExist(err) {
		t.Errorf("expected not exist error, but got: %v", err)
	}

	_, err = service.Get(context.Background(), "not/exist", "not-exist")
	if !os.IsNotExist(err) {
		t.Errorf("expected not exist error, but got: %v", err)
	}

	_, err = service.Toggle(context.Background(), "not/exist", "not-exist", reactions.ToggleRequest{})
	if !os.IsNotExist(err) {
		t.Errorf("expected not exist error, but got: %v", err)
	}

	err = vfsutil.MkdirAll(context.Background(), memFS, "dir/dir", 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Get(context.Background(), "dir", "dir")
	if !os.IsNotExist(err) {
		t.Errorf("expected not exist error, but got: %v", err)
	}
}

type mockUsers struct {
	users.Service
}

func (mockUsers) Get(_ context.Context, user users.UserSpec) (users.User, error) {
	switch {
	case user == users.UserSpec{ID: 1, Domain: "example.org"}:
		return users.User{
			UserSpec: user,
			Login:    "gopher",
			Name:     "Sample Gopher",
			Email:    "gopher@example.org",
		}, nil
	default:
		return users.User{}, fmt.Errorf("user %v not found", user)
	}
}

func (mockUsers) GetAuthenticatedSpec(_ context.Context) (users.UserSpec, error) {
	return users.UserSpec{ID: 1, Domain: "example.org"}, nil
}

func (m mockUsers) GetAuthenticated(ctx context.Context) (users.User, error) {
	userSpec, err := m.GetAuthenticatedSpec(ctx)
	if err != nil {
		return users.User{}, err
	}
	if userSpec.ID == 0 {
		return users.User{}, nil
	}
	return m.Get(ctx, userSpec)
}
