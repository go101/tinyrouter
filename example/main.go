package main

import (
	"fmt"
	"log"
	"net/http"

	tiny "github.com/go101/tinyrouter"
)

func main() {
	makeRoute := func(method, pattern string) tiny.Route {
		f := func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte(pattern))
			w.Write([]byte("   "))
			params := tiny.PathParams(req)
			w.Write([]byte(fmt.Sprintf("params: %v\n", params)))
		}
		return tiny.Route{Method: method, Pattern: pattern, HandleFunc: f}
	}

	routes := []tiny.Route{
		makeRoute("GET", "/organizations/:param1/members/:param2"),
		makeRoute("GET", "/organizations/:paramA/projects/:param2"),

		makeRoute("GET", "/v1/:name/selection"),
		makeRoute("GET", "/v1/user/selection"),

		makeRoute("GET", "/v2/user/selection"),
		makeRoute("GET", "/v2/:name/selection"),
		makeRoute("GET", "/v2/:group/:item/settings"),
		makeRoute("GET", "/v2/:group/:item"),
		makeRoute("GET", "/v2/:id"),

		makeRoute("GET", "/"),
		makeRoute("GET", "/:item"),
	}
	router := tiny.New(&tiny.Config{
		Routes: routes,
		OthersHandleFunc: func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("not found\n"))
		},
	})

	router.Dump() // for debug

	log.Println("Starting service ...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
