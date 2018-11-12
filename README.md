
**NOTE**: if your project supports Go modules, then the import path of this package is `go101.org/tinyrouter`,
so please add the following lines in your project `go.mod` file to use this package:
```
require go101.org/tinyrouter v1.0.1 // this line may be not needed any more since Go 1.12
replace go101.org/tinyrouter => github.com/go101/tinyrouter v1.0.1
```

### What?

TineyRouter is a tiny Go http router supporting custom parameters in paths
(500 lines of code).

The Go package implements an **_O(2k)_** complexity algorithm (usual case) to route HTTP requests.
where **_k_** is the length of a HTTP request path.

### Why?

For a long time, Julien Schmidt's [HttpRouter](https://github.com/julienschmidt/HttpRouter)
is my favorite http router and is used in my many Go projects.
For most cases, HttpRouter works very well.
However, sometimes HttpRouter is some frustrating for [lacking of flexibity](https://github.com/julienschmidt/HttpRouter/search?q=conflicts&type=Issues).
For example, the path patterns in each of the following groups are conflicted with each other in HttpRouter.

```
	// 1
	/organizations/:param1/members/:param2
	/organizations/:abc/projects/:param2

	// 2
	/v1/user/selection
	/v1/:name/selection

	// 3
	/v2/:user/info
	/v2/:user/:group

	// 4
	/v3/user/selection
	/v3/:name

	// 5
	/sub/:group/:item
	/sub/:id
	
	// 6
	/a/b/:c
	/a/:b/c
	/a/:b/:c
	/:a/b/c
	/:a/:b/:c
```

TinyRouter is router implementation between HttpRouter and [gorilla/mux](https://github.com/gorilla/mux),
from both performance and flexibility views.
In practice, for most general cases, TinyRouter is pretty fast.
And, the above routes which don't work in HttpRouter all work fine in TinyRouter.

Like many other router packages, a token in path patterns starting a `:`
is viewed as a parameter. Regexp patterns are not supported by TinyRouter.

An example by using TinyRouter:

```golang
package main

import (
	"fmt"
	"log"
	"net/http"

	tiny "github.com/go101/tinyrouter"
	// tiny "go101.org/tinyrouter" // If your project supports Go modules, please use this line instead.
)

func main() {
	routes := []tiny.Route{
		{
			Method: "GET",
			Pattern: "/design/:user/:slot",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/design/:user/:slot", "user:", params.Value("user"), "slot:", params.Value("slot"))
			},
		},
		{
			Method: "GET",
			Pattern: "/design/:user/:slot/settings/show",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/design/:user/:slot/settings/show", "user:", params.Value("user"), "slot:", params.Value("slot"))
			},
		},
		{
			Method: "POST",
			Pattern: "/design/:user/:slot/settings/update",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/design/:user/:slot/settings/update", "user:", params.Value("user"), "slot:", params.Value("slot"))
			},
		},
		{
			Method: "POST",
			Pattern: "/design/:user/:slot/delete",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/design/:user/:slot/delete", "user:", params.Value("user"), "slot:", params.Value("slot"))
			},
		},
		{
			Method: "GET",
			Pattern: "/design/:uuid",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/design/:uuid", "uuid =", params.Value("uuid"))
			},
		},
		{
			Method: "GET",
			Pattern: "/design/:uuid/stat",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/design/:uuid/stat", "uuid =", params.Value("uuid"))
			},
		},
		{
			Method: "GET",
			Pattern: "/",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintln(w, "/")
			},
		},
		{
			Method: "GET",
			Pattern: "/:sitepage",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/", "sitepage =", params.Value("sitepage"))
			},
		},
	}
	
	config := tiny.Config{
		Routes:           routes,
		OthersHandleFunc: func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintln(w, "other pages")
		},
	}
	
	router := tiny.New(config)

	log.Println("Starting service ...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

Fixed tokens in patterns have higher precedences than parameterized ones.
Left tokens have higher precedences than right ones.
The following patterns are shown by their precedence:
```
1: /a/b/:c
2: /a/:b/c
4: /a/:b/:c
3: /:a/b/c
5: /:a/:b/:c
```
So,
* `/a/b/c` will match the 1st one
* `a/x/c` will match the 2nd one,
* `a/x/y` will match the 2nd one,
* `/x/b/c` will match the 4th one.
* `/y/x/z` will match the 5th one.

The match rules and results are the same as Gorilla/Mux.

### How?

The TinyRouter implementation groups routes:
1. first by request methods.
1. then by number of tokens (or called segments) in path patterns.
1. then (for the 1st segment in patterns), by wildcard (parameterized) or not. Non-wildcard segments are called fixed segments.
1. then for the segments in the fixed group in the last step, group them by their length.
1. for each group with the same token length, sort the segments in it.

(Repeat the last two steps for 2nd, 3rd, ..., segments.)

When a request comes, its URL path will be parsed into tokens (one **k** in **_O(2k + N)_**).
1. The route group with the exact reqest method will be selected.
1. Then the route sub-group (by number of tokens) with the exact number of tokens will be selected.
1. Then, for the 1st token, find the start segment with the same length in the fixed groups
   and start comparing the token with the same-length segments.
   Most `len(token)` bytes will be compared in this comparision.
   If a fixed match is found, then try to find the match for the next token.
   If no matches are found, then try to find the match for next token in the wildcard group.

(Repeat the last step, until a match is found or return without any matches.
Another **k** and the **N** happen in the process.
Some micro-optimizations, by using some one-time built information,
in the process make the usual time complexity become to **_O(2k + N/m)_**.)

For a project with 20 routes per method with a certain number of segments in path,
**_N/m_** would be about 5, whcih is much smaller than **k**, which is about 16-64.
So the usual time complexity of this algorithm is about two times of a radix implementation
(see [the benchmarks](benchmarks/benchmark.md) for details).
The benefit is there are less limits for the route patterns.

