// +build ignore

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/shurcooL/issues"
	"github.com/shurcooL/issues/fs"
	"github.com/shurcooL/issuesapp"
	"github.com/shurcooL/users"
	"golang.org/x/net/webdav"
)

var (
	httpFlag = flag.String("http", ":8080", "Listen for HTTP connections on this address.")
)

func main() {
	flag.Parse()

	var err error
	err = initApp()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Started.")

	err = http.ListenAndServe(*httpFlag, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func initApp() error {
	users := Users{gh: github.NewClient(nil)}
	service, err := fs.NewService(webdav.Dir(filepath.Join(os.Getenv("HOME"), "Dropbox", "Store", "issues")), nil, users)
	if err != nil {
		return err
	}

	opt := issuesapp.Options{
		HeadPre: `<!--link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.0.0-alpha/css/bootstrap.css" media="all" rel="stylesheet" type="text/css" /-->
<style type="text/css">
	body {
		margin: 20px;
		font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
		font-size: 14px;
		line-height: initial;
		color: #373a3c;
	}
	a {
		color: #0275d8;
		text-decoration: none;
	}
	a:focus, a:hover {
		color: #014c8c;
		text-decoration: underline;
	}
	.btn {
		font-size: 11px;
		line-height: 11px;
		border-radius: 4px;
		border: solid #d2d2d2 1px;
		background-color: #fff;
		box-shadow: 0 1px 1px rgba(0, 0, 0, .05);
	}
</style>`,
		BodyPre: `<div style="text-align: right; margin-bottom: 20px; height: 18px; font-size: 12px;">
	{{if .CurrentUser.ID}}
		<a class="topbar-avatar" href="{{.CurrentUser.HTMLURL}}" target="_blank" tabindex=-1
			><img class="topbar-avatar" src="{{.CurrentUser.AvatarURL}}" title="Signed in as {{.CurrentUser.Login}}."
		></a>
		<form method="post" action="/logout" style="display: inline-block; margin-bottom: 0;"><input class="btn" type="submit" value="Sign out"><input type="hidden" name="return" value="{{.BaseURI}}{{.ReqPath}}"></form>
	{{else}}
		<form method="post" action="/login/github" style="display: inline-block; margin-bottom: 0;"><input class="btn" type="submit" value="Sign in via GitHub"><input type="hidden" name="return" value="{{.BaseURI}}{{.ReqPath}}"></form>
	{{end}}
</div>`,
	}
	issuesApp := issuesapp.New(service, users, opt)

	appHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		prefixLen := len("/blog")
		if prefix := req.URL.Path[:prefixLen]; req.URL.Path == prefix+"/" {
			baseURL := prefix
			if req.URL.RawQuery != "" {
				baseURL += "?" + req.URL.RawQuery
			}
			http.Redirect(w, req, baseURL, http.StatusFound)
			return
		}
		req.URL.Path = req.URL.Path[prefixLen:]
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req = req.WithContext(context.WithValue(req.Context(), issuesapp.RepoSpecContextKey, issues.RepoSpec{"github.com/shurcooL/issuesapp"}))
		req = req.WithContext(context.WithValue(req.Context(), issuesapp.BaseURIContextKey, "/blog"))
		issuesApp.ServeHTTP(w, req)
	})
	http.Handle("/blog", appHandler)
	http.Handle("/blog/", appHandler)

	return nil
}

// Users implementats users.Service.
type Users struct {
	gh *github.Client
}

func (s Users) Get(ctx context.Context, user users.UserSpec) (users.User, error) {
	const (
		gh = "github.com"
		tw = "twitter.com"
	)

	switch {
	case user == users.UserSpec{ID: 1924134, Domain: gh}:
		// TODO: Consider using UserSpec{ID: 1, Domain: ds} as well.
		return users.User{
			UserSpec:  user,
			Elsewhere: []users.UserSpec{{ID: 21361484, Domain: tw}},
			Login:     "shurcooL",
			Name:      "Dmitri Shuralyov",
			AvatarURL: "https://dmitri.shuralyov.com/avatar.jpg",
			HTMLURL:   "https://dmitri.shuralyov.com",
			SiteAdmin: true,
		}, nil

	case user.Domain == "github.com":
		ghUser, _, err := s.gh.Users.GetByID(int(user.ID))
		if err != nil {
			return users.User{}, err
		}
		if ghUser.Login == nil || ghUser.AvatarURL == nil || ghUser.HTMLURL == nil {
			return users.User{}, fmt.Errorf("github user missing fields: %#v", ghUser)
		}
		return users.User{
			UserSpec:  user,
			Login:     *ghUser.Login,
			AvatarURL: template.URL(*ghUser.AvatarURL),
			HTMLURL:   template.URL(*ghUser.HTMLURL),
		}, nil

	default:
		return users.User{}, fmt.Errorf("user %v not found", user)
	}
}

func (s Users) GetAuthenticatedSpec(ctx context.Context) (users.UserSpec, error) {
	// TEMP, HACK: Pretend I'm logged in (for testing).
	return users.UserSpec{ID: 1924134, Domain: "github.com"}, nil

	// TEMP, HACK: Pretend I'm logged in as non-admin user.
	//return &users.UserSpec{ID: 4332971, Domain: "github.com"}, nil
}

func (s Users) GetAuthenticated(ctx context.Context) (users.User, error) {
	userSpec, err := s.GetAuthenticatedSpec(ctx)
	if err != nil {
		return users.User{}, err
	}
	if userSpec.ID == 0 {
		return users.User{}, nil
	}
	return s.Get(ctx, userSpec)
}

func (Users) Edit(ctx context.Context, er users.EditRequest) (users.User, error) {
	return users.User{}, errors.New("Edit is not implemented")
}
