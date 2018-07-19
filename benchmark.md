
### Code

[benchmark_test.go](benchmark_test.go)

### Results

```
$ go test -bench=. -benchtime=3s -benchmem
goos: linux
goarch: amd64
pkg: github.com/go101/tinyrouter
Benchmark_HttpRouter_0bytes-4                    	  300000	     16007 ns/op	    4592 B/op	      65 allocs/op
Benchmark_TinyRouter_0bytes-4                    	  200000	     32577 ns/op	   11232 B/op	     117 allocs/op
Benchmark_GorillaMux_0bytes-4                    	   30000	    134893 ns/op	   20746 B/op	     183 allocs/op
Benchmark_HttpRouter_16bytes-4                   	  200000	     18073 ns/op	    4800 B/op	      78 allocs/op
Benchmark_TinyRouter_16bytes-4                   	  200000	     34823 ns/op	   11440 B/op	     130 allocs/op
Benchmark_GorillaMux_16bytes-4                   	   30000	    136889 ns/op	   20947 B/op	     196 allocs/op
Benchmark_HttpRouter_256bytes-4                  	  100000	     52227 ns/op	   13952 B/op	     299 allocs/op
Benchmark_TinyRouter_256bytes-4                  	  100000	     69620 ns/op	   20592 B/op	     351 allocs/op
Benchmark_GorillaMux_256bytes-4                  	   30000	    139576 ns/op	   20952 B/op	     196 allocs/op
Benchmark_HttpRouter_1024bytes-4                 	   30000	    152497 ns/op	   48896 B/op	     949 allocs/op
Benchmark_TinyRouter_1024bytes-4                 	   30000	    171511 ns/op	   55536 B/op	    1001 allocs/op
Benchmark_GorillaMux_1024bytes-4                 	   20000	    294830 ns/op	   65840 B/op	    1069 allocs/op
Benchmark_HttpRouter_8192bytes-4                 	    5000	    953831 ns/op	  380032 B/op	    6812 allocs/op
Benchmark_TinyRouter_8192bytes-4                 	    5000	    963897 ns/op	  386672 B/op	    6864 allocs/op
Benchmark_GorillaMux_8192bytes-4                 	    5000	   1130039 ns/op	  402353 B/op	    6949 allocs/op
Benchmark_HttpRouter_65536bytes-4                	     500	   7195465 ns/op	 2989184 B/op	   53443 allocs/op
Benchmark_TinyRouter_65536bytes-4                	    1000	   7007792 ns/op	 2995826 B/op	   53495 allocs/op
Benchmark_GorillaMux_65536bytes-4                	     500	   7383657 ns/op	 3047288 B/op	   53689 allocs/op
Benchmark_TinyRouter_FlexiblePatterns_0bytes-4   	  200000	     31346 ns/op	   10723 B/op	     120 allocs/op
Benchmark_GorillaMux_FlexiblePatterns_0bytes-4   	   30000	    158160 ns/op	   15694 B/op	     161 allocs/op
```

Note for the second group of benchmarks (the last two) doesn't include HttpRouter,
for HttpRouter will panic on those flexible URL patterns.

From the results, we can find that
* if the HTTP response content length is small, HttpRouter performs best.
TinyRouter uses about two times time of HttpRouter (consistent with theoretical expectation).
GorillaMux uses about eight times time of HttpRouter (quite good for a rich feature router implementation. Much better than I expected).
* the larger the HTTP response content length is,
the smaller performance differences between the three routers.
When the HTTP response content length is larger than 8k,
we can ignore the impact of routers.

### Conclusions

* If your server project needs a high URL pattern flexibility,
then GorillaMux is the best choice.
In particular if your HTTP server is for serving large HTML pages purpose.
For this purpose, router is not an important factor for the performance of your server.
* If your server project needs a medium URL pattern flexibility,
your server may serve a large quantity of small-length content,
and you do care about the the performance of your server,
then TinyRouter is the best choice.
* If the URL pattern flexibility requirement of your server project is low,
and you do care about the the performance of your server,
then HttpRouter is the best choice.

### Other Comparisons

Use standard `http.HandlerFunc` as handler functions?
* HttpRouter: optional
* GorillaMux: Yes
* TinyRouter: Yes

Use standard `http.Request.WithContext` to store parameters?
* HttpRouter: optional
* GorillaMux: optional
* TinyRouter: Yes

Lines of code:
* HttpRouter: 1000
* GorillaMux: 1500
* TinyRouter: 500

Features:
* HttpRouter: limited
* GorillaMux: rich
* TinyRouter: limited
