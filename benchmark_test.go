package tinyrouter

import "io"
import "io/ioutil"
import "net/http"
import "net/http/httptest"
import "strings"
import "testing"

//import TinyRouter "github.com/go101/tinyrouter"
import HttpRouter "github.com/julienschmidt/httprouter"
import GorillaMux "github.com/gorilla/mux"

var _ = ioutil.ReadAll

var requestPatterns = []string{
	"/v1/namespaces/:param0/apps/:param1",
	"/v1/namespaces/:param0/apps/:param1/settings",
	"/v1/namespaces/:param0/apps/:param1/stars",
	"/v1/namespaces/:param0/apps/:param1/stars/by/:param2",
	"/v1/namespaces/:param0/pods/:param1",
	"/v1/namespaces/:param0/pods/:param1/info",
	"/v1/namespaces/:param0/services/:param1",
	"/v1/namespaces/:param0/services/:param1/status",
	"/v1/namespaces/:param0/services/:param1/logs",
	"/v1/namespaces/:param0/services/:param1/logs/:param2",
	"/v1/accounts/:param0",
	"/v1/accounts/:param0/about",
	"/v1/accounts/:param0/settings",
	//"/v1/:uuid", // HttpRouter will panic for this pattern.
}
var requestURLs = []string{
	"/v1/namespaces/google/apps/android",
	"/v1/namespaces/google/apps/android/settings",
	"/v1/namespaces/google/apps/android/stars",
	"/v1/namespaces/google/apps/android/stars/by/trump",
	"/v1/namespaces/google/pods/android",
	"/v1/namespaces/google/pods/:pod/info",
	"/v1/namespaces/google/services/play",
	"/v1/namespaces/google/services/play/status",
	"/v1/namespaces/google/services/play/logs",
	"/v1/namespaces/google/services/play/logs/457a8a5e-89c0-11e8-8893-cf8b2f0abf07",
	"/v1/accounts/trump",
	"/v1/accounts/trump/about",
	"/v1/accounts/trump/settings",
}
var requests []*http.Request

var requestPatterns_2 = []string{
	"/aaaaaaaaaa/bbbbbbbbbb/ccccccccccc/ddddddddddd",
	"/aaaaaaaaaa/bbbbbbbbbb/ccccccccccc/:param0",
	"/aaaaaaaaaa/bbbbbbbbbb/:param0/ddddddddddd",
	"/aaaaaaaaaa/bbbbbbbbbb/:param0/:param1",
	"/aaaaaaaaaa/:param0/ccccccccccc/ddddddddddd",
	"/aaaaaaaaaa/:param0/ccccccccccc/:param1",
	"/aaaaaaaaaa/:param0/:param1/ddddddddddd",
	"/aaaaaaaaaa/:param0/:param1/:param2",
	"/:param0/bbbbbbbbbb/ccccccccccc/ddddddddddd",
	"/:param0/bbbbbbbbbb/ccccccccccc/:param1",
	"/:param0/bbbbbbbbbb/:param1/ddddddddddd",
	"/:param0/bbbbbbbbbb/:param1/:param2",
	"/:param0/:param1/ccccccccccc/ddddddddddd",
	"/:param0/:param1/ccccccccccc/:param2",
	"/:param0/:param1/:param2/ddddddddddd",
	"/:param0/:param1/:param2/:param3",
}
var requestURLs_2 = []string{
	"/aaaaaaaaaa/bbbbbbbbbb/ccccccccccc/ddddddddddd",
	"/aaaaaaaaaa/bbbbbbbbbb/ccccccccccc/xxxxxxxxxxx",
	"/aaaaaaaaaa/bbbbbbbbbb/xxxxxxxxxxx/ddddddddddd",
	"/aaaaaaaaaa/bbbbbbbbbb/xxxxxxxxxxx/yyyyyyyyyyy",
	"/aaaaaaaaaa/xxxxxxxxxx/ccccccccccc/ddddddddddd",
	"/aaaaaaaaaa/xxxxxxxxxx/ccccccccccc/yyyyyyyyyyy",
	"/aaaaaaaaaa/xxxxxxxxxx/yyyyyyyyyyy/ddddddddddd",
	"/aaaaaaaaaa/xxxxxxxxxx/yyyyyyyyyyy/zzzzzzzzzzz",
	"/xxxxxxxxxx/bbbbbbbbbb/ccccccccccc/ddddddddddd",
	"/xxxxxxxxxx/bbbbbbbbbb/ccccccccccc/yyyyyyyyyyy",
	"/xxxxxxxxxx/bbbbbbbbbb/yyyyyyyyyyy/ddddddddddd",
	"/xxxxxxxxxx/bbbbbbbbbb/yyyyyyyyyyy/zzzzzzzzzzz",
	"/xxxxxxxxxx/yyyyyyyyyy/ccccccccccc/ddddddddddd",
	"/xxxxxxxxxx/yyyyyyyyyy/ccccccccccc/zzzzzzzzzzz",
	"/xxxxxxxxxx/yyyyyyyyyy/zzzzzzzzzzz/ddddddddddd",
	"/xxxxxxxxxx/yyyyyyyyyy/zzzzzzzzzzz/wwwwwwwwwww",
}
var requests_2 []*http.Request

func handle(req *http.Request, handler http.Handler) {
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	//resp := w.Result()
	//defer resp.Body.Close()
	//_, _ = ioutil.ReadAll(resp.Body)
}

func write0bytes(w io.Writer) {
}

func write16bytes(w io.Writer) {
	w.Write([]byte{'h', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', '!', ' ', ' ', ' '})
}

func write256bytes(w io.Writer) {
	for i := 0; i < 16; i++ {
		write16bytes(w)
	}
}

func write1024bytes(w io.Writer) {
	for i := 0; i < 64; i++ {
		write16bytes(w)
	}
}

func write8192bytes(w io.Writer) {
	for i := 0; i < 512; i++ {
		write16bytes(w)
	}
}

func write65536bytes(w io.Writer) {
	for i := 0; i < 4096; i++ {
		write16bytes(w)
	}
}

func handlerTinyRouter(f func(io.Writer)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		params := /*TinyRouter.*/ PathParams(req)
		_, _, _ = params.Value("param0"), params.Value("param1"), params.Value("param2")
		w.WriteHeader(http.StatusOK)
		f(w)
	}
}
func handlerHttpRouter(f func(io.Writer)) func(w http.ResponseWriter, req *http.Request, params HttpRouter.Params) {
	return func(w http.ResponseWriter, req *http.Request, params HttpRouter.Params) {
		_, _, _ = params.ByName("param0"), params.ByName("param1"), params.ByName("param2")
		w.WriteHeader(http.StatusOK)
		f(w)
	}
}
func handlerGorillaMux(f func(io.Writer)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		params := GorillaMux.Vars(req)
		_, _, _ = params["param0"], params["param1"], params["param2"]
		w.WriteHeader(http.StatusOK)
		f(w)
	}
}

var tinyRouter0, tinyRouter16, tinyRouter256, tinyRouter1024, tinyRouter8192, tinyRouter65536, tinyRouter0_b * /*TinyRouter.*/ TinyRouter
var httpRouter0, httpRouter16, httpRouter256, httpRouter1024, httpRouter8192, httpRouter65536 *HttpRouter.Router
var gorillaRouter0, gorillaRouter16, gorillaRouter256, gorillaRouter1024, gorillaRouter8192, gorillaRouter65536, gorillaRouter0_b *GorillaMux.Router

func init() {
	// ...
	requests = make([]*http.Request, 0, len(requestURLs))
	for _, path := range requestURLs {
		requests = append(requests, httptest.NewRequest("GET", "http://example.com"+path, nil))
	}
	requests_2 = make([]*http.Request, 0, len(requestURLs))
	for _, path := range requestURLs_2 {
		requests_2 = append(requests_2, httptest.NewRequest("GET", "http://example.com"+path, nil))
	}

	// TinyRouter
	tinyroutes0 := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	tinyroutes16 := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	tinyroutes256 := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	tinyroutes1024 := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	tinyroutes8192 := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	tinyroutes65536 := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	for _, pattern := range requestPatterns {
		tinyroutes0 = append(tinyroutes0 /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write0bytes),
		})
		tinyroutes16 = append(tinyroutes16 /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write16bytes),
		})
		tinyroutes256 = append(tinyroutes256 /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write256bytes),
		})
		tinyroutes1024 = append(tinyroutes1024 /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write1024bytes),
		})
		tinyroutes8192 = append(tinyroutes8192 /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write8192bytes),
		})
		tinyroutes65536 = append(tinyroutes65536 /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write65536bytes),
		})
	}
	tinyRouter0 = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes0})
	tinyRouter16 = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes16})
	tinyRouter256 = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes256})
	tinyRouter1024 = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes1024})
	tinyRouter8192 = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes8192})
	tinyRouter65536 = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes65536})

	tinyroutes0_b := make([] /*TinyRouter.*/ Route, 0, len(requestPatterns))
	for _, pattern := range requestPatterns_2 {
		tinyroutes0_b = append(tinyroutes0_b /*TinyRouter.*/, Route{
			Method:     "GET",
			Pattern:    pattern,
			HandleFunc: handlerTinyRouter(write0bytes),
		})
	}
	tinyRouter0_b = /*TinyRouter.*/ New( /*TinyRouter.*/ Config{Routes: tinyroutes0_b})

	// HttpRouter
	httpRouter0 = HttpRouter.New()
	httpRouter16 = HttpRouter.New()
	httpRouter256 = HttpRouter.New()
	httpRouter1024 = HttpRouter.New()
	httpRouter8192 = HttpRouter.New()
	httpRouter65536 = HttpRouter.New()
	for _, pattern := range requestPatterns {
		httpRouter0.GET(pattern, handlerHttpRouter(write0bytes))
		httpRouter16.GET(pattern, handlerHttpRouter(write16bytes))
		httpRouter256.GET(pattern, handlerHttpRouter(write256bytes))
		httpRouter1024.GET(pattern, handlerHttpRouter(write1024bytes))
		httpRouter8192.GET(pattern, handlerHttpRouter(write8192bytes))
		httpRouter65536.GET(pattern, handlerHttpRouter(write65536bytes))
	}

	// GorillaMux
	gorillaRouter0 = GorillaMux.NewRouter()
	gorillaRouter16 = GorillaMux.NewRouter()
	gorillaRouter256 = GorillaMux.NewRouter()
	gorillaRouter1024 = GorillaMux.NewRouter()
	gorillaRouter8192 = GorillaMux.NewRouter()
	gorillaRouter65536 = GorillaMux.NewRouter()
	for _, pattern := range requestPatterns {
		tokens := strings.Split(pattern, "/")
		pattern = ""
		for i, t := range tokens {
			if i := strings.IndexByte(t, ':'); i >= 0 {
				pattern += "{" + t[1:] + "}"
			} else {
				pattern += t
			}
			if i+1 < len(tokens) {
				pattern += "/"
			}
		}
		gorillaRouter0.HandleFunc(pattern, handlerGorillaMux(write0bytes)).Methods("GET")
		gorillaRouter16.HandleFunc(pattern, handlerGorillaMux(write16bytes)).Methods("GET")
		gorillaRouter256.HandleFunc(pattern, handlerGorillaMux(write16bytes)).Methods("GET")
		gorillaRouter1024.HandleFunc(pattern, handlerGorillaMux(write1024bytes)).Methods("GET")
		gorillaRouter8192.HandleFunc(pattern, handlerGorillaMux(write8192bytes)).Methods("GET")
		gorillaRouter65536.HandleFunc(pattern, handlerGorillaMux(write65536bytes)).Methods("GET")
	}

	gorillaRouter0_b = GorillaMux.NewRouter()
	for _, pattern := range requestPatterns_2 {
		tokens := strings.Split(pattern, "/")
		pattern = ""
		for i, t := range tokens {
			if i := strings.IndexByte(t, ':'); i >= 0 {
				pattern += "{" + t[1:] + "}"
			} else {
				pattern += t
			}
			if i+1 < len(tokens) {
				pattern += "/"
			}
		}
		gorillaRouter0_b.HandleFunc(pattern, handlerGorillaMux(write0bytes)).Methods("GET")
	}
}

func Benchmark_HttpRouter_0bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, httpRouter0)
		}
	}
}

func Benchmark_TinyRouter_0bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter0)
		}
	}
}

func Benchmark_GorillaMux_0bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter0)
		}
	}
}

func Benchmark_HttpRouter_16bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, httpRouter16)
		}
	}
}

func Benchmark_TinyRouter_16bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter16)
		}
	}
}

func Benchmark_GorillaMux_16bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter16)
		}
	}
}

func Benchmark_HttpRouter_256bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, httpRouter256)
		}
	}
}

func Benchmark_TinyRouter_256bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter256)
		}
	}
}

func Benchmark_GorillaMux_256bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter256)
		}
	}
}

func Benchmark_HttpRouter_1024bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, httpRouter1024)
		}
	}
}

func Benchmark_TinyRouter_1024bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter1024)
		}
	}
}

func Benchmark_GorillaMux_1024bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter1024)
		}
	}
}

func Benchmark_HttpRouter_8192bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, httpRouter8192)
		}
	}
}

func Benchmark_TinyRouter_8192bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter8192)
		}
	}
}

func Benchmark_GorillaMux_8192bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter8192)
		}
	}
}

func Benchmark_HttpRouter_65536bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, httpRouter65536)
		}
	}
}

func Benchmark_TinyRouter_65536bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter65536)
		}
	}
}

func Benchmark_GorillaMux_65536bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter65536)
		}
	}
}

// flexible patterns

func Benchmark_TinyRouter_FlexiblePatterns_0bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, tinyRouter0_b)
		}
	}
}

func Benchmark_GorillaMux_FlexiblePatterns_0bytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			handle(req, gorillaRouter0_b)
		}
	}
}
