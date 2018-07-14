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
		urlPath        string
		expectedParams map[string]string
		expectedValues []string
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
			ps := PathParams(r)
			params, values := map[string]string{}, []string(nil)
			for i := 0; i < len(ps.kvs); i += 2 {
				params[ps.kvs[i]] = ps.kvs[i+1]
				values = append(values, ps.kvs[i+1])
			}

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
				Pattern: "/organizations/:param1/projects/:param2",
			},
			requests: []requestCase{
				{
					urlPath:        "/organizations/Google/projects/Android",
					expectedParams: map[string]string{"param1": "Google", "param2": "Android"},
					expectedValues: []string{"Google", "Android"},
				},
				{
					urlPath:        "/organizations/Apple/projects/iPhone",
					expectedParams: map[string]string{"param1": "Apple", "param2": "iPhone"},
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
	}

	// ...
	routes := []Route{}
	for _, rc := range routeCases {
		route := rc.route
		route.HandleFunc = buildHandler(rc)
		routes = append(routes, route)
	}
	router := New(&Config{Routes: routes})

	// ...
	for _, rc := range routeCases {
		for _, reqc := range rc.requests {
			req := httptest.NewRequest(rc.route.Method, "http://example.com"+reqc.urlPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)
			var resc responseCase
			_ = json.Unmarshal(body, &resc)
			if resc.Method != rc.route.Method {
				t.Errorf("method not match: %s : %s %s", resc.Method, rc.route.Method, rc.route.Pattern)
			}
			if resc.Pattern != rc.route.Pattern {
				t.Errorf("pattern not match: %s : %s %s", resc.Pattern, rc.route.Method, rc.route.Pattern)
			}
			for k, v := range resc.Params {
				if v != reqc.expectedParams[k] {
					t.Errorf("param value not match: [%s] / %s / %s : %s %s", k, v, reqc.expectedParams[k], rc.route.Method, rc.route.Pattern)
				}
			}
			for k, v := range reqc.expectedParams {
				if v != resc.Params[k] {
					t.Errorf("param value not match: [%s] / %s / %s : %s %s", k, resc.Params[k], v, rc.route.Method, rc.route.Pattern)
				}
			}
			if len(resc.Values) != len(reqc.expectedValues) {
				t.Errorf("number of params not match: %d / %d : %s %s", len(resc.Values), len(reqc.expectedValues), rc.route.Method, rc.route.Pattern)
			}
			for i, v := range resc.Values {
				if reqc.expectedValues[i] != v {
					t.Errorf("param value not match: [%d] / %s / %s : %s %s", i, reqc.expectedValues[i], v, rc.route.Method, rc.route.Pattern)
				}
			}
		}
	}
}
