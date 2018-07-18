package tinyrouter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTinyRouter(t *testing.T) {

	type requestCase struct {
		urlPath         string
		expectedParams  map[string]string
		expectedValues  []string
		expectedPattern string // for invert testing. If this is set, route.Pattern must be unset.
	}

	type routeCase struct {
		route    Route
		requests []requestCase
	}

	type responseCase struct {
		Method  string
		Pattern string
		Params  map[string]string
		Values  []string
	}

	buildHandler := func(rc routeCase) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			params, values := PathParams(r).ToMapAndSlice()
			res := responseCase{
				Method:  r.Method,
				Pattern: rc.route.Pattern,
				Params:  params,
				Values:  values,
			}
			data, _ := json.Marshal(res)
			w.Write(data)
		}
	}

	// ...
	var routeCases = []routeCase{
		{
			route: Route{
				Method:  "GET",
				Pattern: "/v1/projects/:project/apps/:app",
			},
			requests: []requestCase{
				{
					urlPath:        "/v1/projects/k8s/apps/controller",
					expectedParams: map[string]string{"project": "k8s", "app": "controller"},
					expectedValues: []string{"k8s", "controller"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/:item",
			},
			requests: []requestCase{
				{
					urlPath:        "/about",
					expectedParams: map[string]string{"item": "about"},
					expectedValues: []string{"about"},
				},
				{
					urlPath:        "/sitemap",
					expectedParams: map[string]string{"item": "sitemap"},
					expectedValues: []string{"sitemap"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/aaabbb",
			},
			requests: []requestCase{
				{
					urlPath:        "/aaabbb",
					expectedParams: map[string]string{},
					expectedValues: []string{},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/cccddd",
			},
			requests: []requestCase{
				{
					urlPath:        "/cccddd",
					expectedParams: map[string]string{},
					expectedValues: []string{},
				},
			},
		},
		{
			route: Route{
				Method: "GET",
			},
			requests: []requestCase{
				{
					urlPath:         "/aaaddd",
					expectedParams:  map[string]string{"item": "aaaddd"},
					expectedValues:  []string{"aaaddd"},
					expectedPattern: "/:item",
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/accounts/admin/info",
			},
			requests: []requestCase{
				{
					urlPath:        "/accounts/admin/info",
					expectedParams: map[string]string{},
					expectedValues: []string{},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/accounts/:name/info",
			},
			requests: []requestCase{
				{
					urlPath:        "/accounts/Alice/info",
					expectedParams: map[string]string{"name": "Alice"},
					expectedValues: []string{"Alice"},
				},
				{
					urlPath:        "/accounts/zhang/info",
					expectedParams: map[string]string{"name": "zhang"},
					expectedValues: []string{"zhang"},
				},
			},
		},
		{
			route: Route{
				Method:  "POST",
				Pattern: "/organizations/:param1/members/:param2",
			},
			requests: []requestCase{
				{
					urlPath:        "/organizations/Google/members/Wang",
					expectedParams: map[string]string{"param1": "Google", "param2": "Wang"},
					expectedValues: []string{"Google", "Wang"},
				},
				{
					urlPath:        "/organizations/Apple/members/Yang",
					expectedParams: map[string]string{"param1": "Apple", "param2": "Yang"},
					expectedValues: []string{"Apple", "Yang"},
				},
			},
		},
		{
			route: Route{
				Method:  "POST",
				Pattern: "/organizations/:/projects/:param2",
			},
			requests: []requestCase{
				{
					urlPath:        "/organizations/Google/projects/Android",
					expectedParams: map[string]string{"param2": "Android"},
					expectedValues: []string{"Google", "Android"},
				},
				{
					urlPath:        "/organizations/Apple/projects/iPhone",
					expectedParams: map[string]string{"param2": "iPhone"},
					expectedValues: []string{"Apple", "iPhone"},
				},
			},
		},
		{
			route: Route{
				Method:  "PUT",
				Pattern: "/v2/foo/bar",
			},
			requests: []requestCase{
				{
					urlPath:        "/v2/foo/bar",
					expectedParams: map[string]string{},
					expectedValues: []string{},
				},
			},
		},
		{
			route: Route{
				Method:  "PUT",
				Pattern: "/v2/:foo/bar",
			},
			requests: []requestCase{
				{
					urlPath:        "/v2/abc/bar",
					expectedParams: map[string]string{"foo": "abc"},
					expectedValues: []string{"abc"},
				},
			},
		},
		{
			route: Route{
				Method:  "PUT",
				Pattern: "/v2/foo/:bar",
			},
			requests: []requestCase{
				{
					urlPath:        "/v2/foo/xyz",
					expectedParams: map[string]string{"bar": "xyz"},
					expectedValues: []string{"xyz"},
				},
			},
		},
		{
			route: Route{
				Method:  "PUT",
				Pattern: "/v2/fo/:bar",
			},
			requests: []requestCase{
				{
					urlPath:        "/v2/fo/xyz",
					expectedParams: map[string]string{"bar": "xyz"},
					expectedValues: []string{"xyz"},
				},
			},
		},
		{
			route: Route{
				Method:  "PUT",
				Pattern: "/v2/:foo/:bar",
			},
			requests: []requestCase{
				{
					urlPath:        "/v2/zhang/san",
					expectedParams: map[string]string{"foo": "zhang", "bar": "san"},
					expectedValues: []string{"zhang", "san"},
				},
				{
					urlPath:        "/v2/John/Smith",
					expectedParams: map[string]string{"foo": "John", "bar": "Smith"},
					expectedValues: []string{"John", "Smith"},
				},
			},
		},
		{
			route: Route{
				Method:  "PUT",
				Pattern: "/v2/:who",
			},
			requests: []requestCase{
				{
					urlPath:        "/v2/zhang",
					expectedParams: map[string]string{"who": "zhang"},
					expectedValues: []string{"zhang"},
				},
				{
					urlPath:        "/v2/John",
					expectedParams: map[string]string{"who": "John"},
					expectedValues: []string{"John"},
				},
			},
		},

		{
			route: Route{
				Method:  "GET",
				Pattern: "/a/b/c/d/:x",
			},
			requests: []requestCase{
				{
					urlPath:        "/a/b/c/d/123",
					expectedParams: map[string]string{"x": "123"},
					expectedValues: []string{"123"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/a/b/:y/d/e",
			},
			requests: []requestCase{
				{
					urlPath:        "/a/b/123/d/e",
					expectedParams: map[string]string{"y": "123"},
					expectedValues: []string{"123"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/:z/b/c/d/e",
			},
			requests: []requestCase{
				{
					urlPath:        "/123/b/c/d/e",
					expectedParams: map[string]string{"z": "123"},
					expectedValues: []string{"123"},
				},
			},
		},

		{
			route: Route{
				Method:  "GET",
				Pattern: "/a123456789/b123456789/c123456789/d123000000",
			},
			requests: []requestCase{
				{
					urlPath:        "/a123456789/b123456789/c123456789/d123000000",
					expectedParams: map[string]string{},
					expectedValues: []string{},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/a123456789/b123456789/:parameter/d123000000",
			},
			requests: []requestCase{
				{
					urlPath:        "/a123456789/b123456789/x/d123000000",
					expectedParams: map[string]string{"parameter": "x"},
					expectedValues: []string{"x"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/a123456789/:parameter/c123456789/d123456789",
			},
			requests: []requestCase{
				{
					urlPath:        "/a123456789/y/c123456789/d123456789",
					expectedParams: map[string]string{"parameter": "y"},
					expectedValues: []string{"y"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/:parameter/b123456789/c123456789/d123456789",
			},
			requests: []requestCase{
				{
					urlPath:        "/z/b123456789/c123456789/d123456789",
					expectedParams: map[string]string{"parameter": "z"},
					expectedValues: []string{"z"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/:parameter/b123456789/:an_option/d123456789",
			},
			requests: []requestCase{
				{
					urlPath:        "/m/b123456789/n/d123456789",
					expectedParams: map[string]string{"parameter": "m", "an_option": "n"},
					expectedValues: []string{"m", "n"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/:parameter/:an_option/c123456789/d123456789",
			},
			requests: []requestCase{
				{
					urlPath:        "/p/q/c123456789/d123456789",
					expectedParams: map[string]string{"parameter": "p", "an_option": "q"},
					expectedValues: []string{"p", "q"},
				},
			},
		},
		{
			route: Route{
				Method:  "GET",
				Pattern: "/:parameter/:an_option/:_whatever/d123456789",
			},
			requests: []requestCase{
				{
					urlPath:        "/h/i/j/d123456789",
					expectedParams: map[string]string{"parameter": "h", "an_option": "i", "_whatever": "j"},
					expectedValues: []string{"h", "i", "j"},
				},
			},
		},
	}

	// ...
	routes := []Route{}
	for _, rc := range routeCases {
		route := rc.route
		if route.Pattern != "" {
			route.HandleFunc = buildHandler(rc)
			routes = append(routes, route)
		}
	}
	router := New(Config{Routes: routes})

	// t.Log(router.DumpInfo())

	// ...
OuterMost:
	for _, rc := range routeCases {
		for _, reqc := range rc.requests {
			req := httptest.NewRequest(rc.route.Method, "http://example.com"+reqc.urlPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			var resc responseCase
			_ = json.Unmarshal(body, &resc)
			if resc.Method != rc.route.Method {
				t.Errorf("method not match: %s : %s %s", resc.Method, rc.route.Method, rc.route.Pattern)
				break OuterMost
			}
			if reqc.expectedPattern != "" && resc.Pattern != reqc.expectedPattern {
				t.Errorf("pattern not match (1): %s : %s %s", resc.Pattern, rc.route.Method, reqc.expectedPattern)
				break OuterMost
			}
			if rc.route.Pattern != "" && resc.Pattern != rc.route.Pattern {
				t.Errorf("pattern not match (2): %s : %s %s", resc.Pattern, rc.route.Method, rc.route.Pattern)
				break OuterMost
			}
			for k, v := range resc.Params {
				if v != reqc.expectedParams[k] {
					t.Errorf("param value not match: [%s] / %s / %s : %s %s", k, v, reqc.expectedParams[k], rc.route.Method, rc.route.Pattern)
					break OuterMost
				}
			}
			for k, v := range reqc.expectedParams {
				if v != resc.Params[k] {
					t.Errorf("param value not match: [%s] / %s / %s : %s %s", k, resc.Params[k], v, rc.route.Method, rc.route.Pattern)
					break OuterMost
				}
			}
			if len(resc.Values) != len(reqc.expectedValues) {
				t.Errorf("number of params not match: %d / %d : %s %s", len(resc.Values), len(reqc.expectedValues), rc.route.Method, rc.route.Pattern)
				break OuterMost
			}
			for i, v := range resc.Values {
				if reqc.expectedValues[i] != v {
					t.Errorf("param value not match: [%d] / %s / %s : %s %s", i, reqc.expectedValues[i], v, rc.route.Method, rc.route.Pattern)
					break OuterMost
				}
			}
		}
	}
}
