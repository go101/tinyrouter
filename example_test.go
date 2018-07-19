package tinyrouter_test

import (
	"fmt"
	"log"
	"net/http"

	tiny "github.com/go101/tinyrouter"
)

func Example() {
	routes := []tiny.Route{
		{
			Method: "GET",
			Pattern: "/a/b/:c",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/a/b/:c", "c =", params.Value("c"))
			},
		},
		{
			Method: "GET",
			Pattern: "/a/:b/c",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/a/:b/c", "b =", params.Value("b"))
			},
		},
		{
			Method: "GET",
			Pattern: "/a/:b/:c",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/a/:b/:c", "b =", params.Value("b"), "c =", params.Value("c"))
			},
		},
		{
			Method: "GET",
			Pattern: "/:a/b/c",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/:a/b/c", "a =", params.Value("a"))
			},
		},
		{
			Method: "GET",
			Pattern: "/:a/:b/:c",
			HandleFunc: func(w http.ResponseWriter, req *http.Request) {
				params := tiny.PathParams(req)
				fmt.Fprintln(w, "/:a/:b/:c", "a =", params.Value("a"), "b =", params.Value("b"), "c =", params.Value("c"))
			},
		},
	}
	
	router := tiny.New(tiny.Config{Routes: routes})

	log.Println("Starting service ...")
	log.Fatal(http.ListenAndServe(":8080", router))
	
	/*
	$ curl localhost:8080/a/b/c
	/a/b/:c c = c
	$ curl localhost:8080/a/x/c
	/a/:b/c b = x
	$ curl localhost:8080/a/x/y
	/a/:b/:c b = x c = y
	$ curl localhost:8080/x/b/c
	/:a/b/c a = x
	$ curl localhost:8080/x/y/z
	/:a/:b/:c a = x b = y c = z
	*/
}

