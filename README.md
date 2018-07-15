A tiny Go http router supporting custom parameters in paths.

The Go package implements an **_O(2k+N)_** complexity algorithm (worst case) to route HTTP requests.
where **_k_** is the length of a HTTP request path and **_N_** is the number of routes to be matched.
For general cases, the real complexity is **_O(2k+N/m)_**, where **_m_** is at the level of ten.

### Why

For a long time, Julien Schmidt's [HttpRouter](https://github.com/julienschmidt/HttpRouter)
is my favorite http router and is used in my many Go projects.
For most cases, HttpRouter works very well.
However, sometimes HttpRouter is some frustrating for [lacking of flexibity](https://github.com/julienschmidt/HttpRouter/search?q=conflicts&type=Issues).
For example, the following route groups don't work at the same time in HttpRouter.

```golang
	router := HttpRouter.New()

	// 1
	router.GET("/organizations/:param1/members/:param2", handle)
	router.GET("/organizations/:abc/projects/:param2", handle)

	// 2
	router.GET("/v1/user/selection", handle)
	router.GET("/v1/:name/selection", handle)

	// 3
	router.GET("/v2/:user/info", handle)
	router.GET("/v2/:user/:group", handle)

	// 4
	router.GET("/v3/user/selection", handle)
	router.GET("/v3/:name", handle)

	// 5
	router.GET("/sub/:group/:item", handle)
	router.GET("/sub/:id", handle)
```

TinyRouter is router implementation between HttpRouter and [gorilla/mux](https://github.com/gorilla/mux),
from both performance (for worst case in theory) and flexibity views.
In practice, for most general cases, TinyRouter is pretty fast.
And, the above routes which don't work in HttpRouter all work fine in TinyRouter.

Partically matched parameters are not supported.

An example by using TinyRouter:

```golang
package main

import (
	"fmt"
	"log"
	"net/http"

	tiny "github.com/go101/tinyrouter"
)

func main() {
	routes := []tiny.Route{
		{
			Method: "GET",
			Pattern: "/organizations/:org/members/:member",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "org: %s, member: %s\n", params.Value("org"), params.Value("member"))
			},
		},
		{
			Method: "GET",
			Pattern: "/organizations/:org/projects/:project",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "org: %s, project: %s\n", params.Value("org"), params.Value("project"))
			},
		},

		{
			Method: "GET",
			Pattern: "/v1/:name/info",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "info of %s (v1)\n", params.Value("name"))
			},
		},
		{
			Method: "GET",
			Pattern: "/v1/website/info",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(w, "website info\n")
			},
		},

		{
			Method: "GET",
			Pattern: "/v2/:name/info",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "info of %s (v2)\n", params.Value("name"))
			},
		},
		{
			Method: "GET",
			Pattern: "/v2/:group/:item",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "grooup: %s, item: %s\n", params.Value("group"), params.Value("item"))
			},
		},
		{
			Method: "GET",
			Pattern: "/v2/:id",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "id: %s(v2)\n", params.Value("id"))
			},
		},


		{
			Method: "GET",
			Pattern: "/",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(w, "home\n")
			},
		},
		{
			Method: "GET",
			Pattern: "/:item",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintf(w, "item: %s\n", params.Value("item"))
			},
		},
	}
	
	router := tiny.New(tiny.Config{Routes: routes})

	log.Println("Starting service ...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

### How

The TinyRouter implementation groups routes:
1. first by number of tokens (or called segments) in path patterns.
1. then by request methods. (This step may be exchanged with the first step in the future versions.)
1. then (for the 1st segment in patterns), by wildcard (parameterized) or not. Non-wildcard segments are called fixed segments.
1. then for the segments in the fixed group in the last step, group them by their length.
1. for each group with the same token length, sort the segments in it.

(Repeat the last steps for 2nd, 3rd, ..., segments.)

When a request comes, its URL path will be parsed into tokens (one **k** in **_O(2k+N)_**).
1. The route group (by number of tokens) with the exact number of tokens will be selected.
1. Then the route sub-group with the exact reqest method will be selected.
1. Then, for the 1st token, find the start segment with the same length in the fixed groups.
   If all the fixed segments with the same length don't match, the try to find the match for next toekn in the wildcard group.

(Repeat the last step, until a match is found or return without any matches.
Another **k** and the **N** happen in the process.
Some micro-optimizations in the process make the common time complexity become to **_O(2k+N/m)_**.)



