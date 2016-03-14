// +build ignore

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/fsissues"
	"github.com/shurcooL/issuesapp"
	"github.com/shurcooL/issuesapp/common"
	"github.com/shurcooL/users"
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/tracker/issues"
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
	users := users.Static{}
	service, err := fs.NewService(filepath.Join(os.Getenv("HOME"), "Dropbox", "Store", "issues"), users)
	if err != nil {
		return err
	}

	opt := issuesapp.Options{
		Context:  func(req *http.Request) context.Context { return context.TODO() },
		RepoSpec: func(req *http.Request) issues.RepoSpec { return issues.RepoSpec{"apps/tracker"} },
		BaseURI:  func(req *http.Request) string { return "/blog" },
		BaseState: func(req *http.Request) issuesapp.BaseState {
			reqPath := req.URL.Path
			if reqPath == "/" {
				reqPath = ""
			}
			return issuesapp.BaseState{
				State: common.State{
					BaseURI: "/blog",
					ReqPath: reqPath,
				},
			}
		},
		HeadPre: `<!--link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.0.0-alpha/css/bootstrap.css" media="all" rel="stylesheet" type="text/css" /-->
<style type="text/css">
	body {
		font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
		font-size: 14px;
		line-height: initial;
		margin: 20px;
	}
	.btn {
		font-size: 14px;
	}
</style>`,
	}
	issuesApp := issuesapp.New(service, opt)

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
		issuesApp.ServeHTTP(w, req)
	})
	http.Handle("/blog", appHandler)
	http.Handle("/blog/", appHandler)

	return nil
}
