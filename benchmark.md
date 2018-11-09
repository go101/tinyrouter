
### Code

[benchmark_test.go](benchmark_test.go)

### Results

```
$ go test -bench=. -benchmem -benchtime=3s
goos: linux
goarch: amd64
pkg: github.com/go101/tinyrouter
Benchmark_TinyRouter_Void-4                      	  200000	     24136 ns/op	    7800 B/op	      78 allocs/op
Benchmark_HttpRouter_Void-4                      	  500000	      7285 ns/op	    1160 B/op	      26 allocs/op
Benchmark_GorillaMux_Void-4                      	   30000	    124647 ns/op	   16936 B/op	     143 allocs/op
Benchmark_TrieMux_Void-4                         	  200000	     19228 ns/op	    5096 B/op	      52 allocs/op
Benchmark_ChiRouter_Void-4                       	  200000	     24352 ns/op	    5721 B/op	      52 allocs/op
Benchmark_TinyRouter_0bytes-4                    	  200000	     31769 ns/op	   11232 B/op	     117 allocs/op
Benchmark_HttpRouter_0bytes-4                    	  300000	     16479 ns/op	    4592 B/op	      65 allocs/op
Benchmark_GorillaMux_0bytes-4                    	   30000	    138028 ns/op	   20368 B/op	     182 allocs/op
Benchmark_TrieMux_0bytes-4                       	  200000	     27597 ns/op	    8528 B/op	      91 allocs/op
Benchmark_ChiRouter_0bytes-4                     	  200000	     34292 ns/op	    9154 B/op	      91 allocs/op
Benchmark_TinyRouter_16bytes-4                   	  200000	     32740 ns/op	   11232 B/op	     117 allocs/op
Benchmark_HttpRouter_16bytes-4                   	  300000	     17331 ns/op	    4592 B/op	      65 allocs/op
Benchmark_GorillaMux_16bytes-4                   	   30000	    139416 ns/op	   20368 B/op	     182 allocs/op
Benchmark_TrieMux_16bytes-4                      	  200000	     28259 ns/op	    8528 B/op	      91 allocs/op
Benchmark_ChiRouter_16bytes-4                    	  200000	     34339 ns/op	    9153 B/op	      91 allocs/op
Benchmark_TinyRouter_256bytes-4                  	  100000	     55050 ns/op	   17264 B/op	     143 allocs/op
Benchmark_HttpRouter_256bytes-4                  	  100000	     39026 ns/op	   10624 B/op	      91 allocs/op
Benchmark_GorillaMux_256bytes-4                  	   30000	    139608 ns/op	   20368 B/op	     182 allocs/op
Benchmark_TrieMux_256bytes-4                     	  100000	     52348 ns/op	   14560 B/op	     117 allocs/op
Benchmark_ChiRouter_256bytes-4                   	  100000	     60024 ns/op	   15187 B/op	     117 allocs/op
Benchmark_TinyRouter_1024bytes-4                 	   50000	    112967 ns/op	   42224 B/op	     169 allocs/op
Benchmark_HttpRouter_1024bytes-4                 	   50000	     97225 ns/op	   35584 B/op	     117 allocs/op
Benchmark_GorillaMux_1024bytes-4                 	   20000	    239795 ns/op	   51360 B/op	     234 allocs/op
Benchmark_TrieMux_1024bytes-4                    	   50000	    111293 ns/op	   39520 B/op	     143 allocs/op
Benchmark_ChiRouter_1024bytes-4                  	   50000	    120545 ns/op	   40153 B/op	     143 allocs/op
Benchmark_TinyRouter_8192bytes-4                 	   10000	    532320 ns/op	  280176 B/op	     208 allocs/op
Benchmark_HttpRouter_8192bytes-4                 	   10000	    502834 ns/op	  273536 B/op	     156 allocs/op
Benchmark_GorillaMux_8192bytes-4                 	   10000	    681497 ns/op	  289312 B/op	     273 allocs/op
Benchmark_TrieMux_8192bytes-4                    	   10000	    524041 ns/op	  277472 B/op	     182 allocs/op
Benchmark_ChiRouter_8192bytes-4                  	   10000	    547148 ns/op	  278153 B/op	     182 allocs/op
Benchmark_TinyRouter_65536bytes-4                	    2000	   3526160 ns/op	 2143857 B/op	     247 allocs/op
Benchmark_HttpRouter_65536bytes-4                	    2000	   3481502 ns/op	 2137217 B/op	     195 allocs/op
Benchmark_GorillaMux_65536bytes-4                	    1000	   3760136 ns/op	 2152994 B/op	     312 allocs/op
Benchmark_TrieMux_65536bytes-4                   	    2000	   3558243 ns/op	 2141153 B/op	     221 allocs/op
Benchmark_ChiRouter_65536bytes-4                 	    2000	   3572608 ns/op	 2142220 B/op	     223 allocs/op
Benchmark_TinyRouter_FlexiblePatterns_0bytes-4   	  100000	     37425 ns/op	   12336 B/op	     140 allocs/op
Benchmark_GorillaMux_FlexiblePatterns_0bytes-4   	   20000	    193210 ns/op	   24800 B/op	     223 allocs/op
Benchmark_TrieMux_FlexiblePatterns_0bytes-4      	  200000	     35188 ns/op	   10160 B/op	     110 allocs/op
Benchmark_ChiRouter_FlexiblePatterns_0bytes-4    	  100000	     42719 ns/op	   11268 B/op	     112 allocs/op
```

Note: the last group of benchmarks doesn't include HttpRouter,
for HttpRouter will panic on those flexible URL patterns.

From the results, we can find that
* (from the first group of benchmarks), HttpRouter performs best for pure router function.
* (from the middle groups of benchmarks), with more and more workload for a single request, the advantage of HttpRouter becomes smaller and smaller.

### Conclusions

* If your server project needs a high URL pattern flexibility, select any route libraries used in the above benchmarks except HttpRouter.
* If the average response size is large than 8k bytes, any router libraries used in the above benchmarks are capable.
* If  the average response size is very small and you don't care about the URL pattern flexibility, HttpRouter may be the best choice.

### Other Comparisons

Use standard `http.HandlerFunc` as handler functions defaultly?
* TinyRouter: Yes
* HttpRouter: No
* GorillaMux: Yes
* TrieMux: No
* ChiRouter: Yes

Use standard `http.Request.WithContext` to store parameters defaultly?
* TinyRouter: Yes
* HttpRouter: No
* GorillaMux: No
* TrieMux: No
* ChiRouter: Yes

Lines of code:
* TinyRouter: 500
* HttpRouter: 1000
* GorillaMux: 1500
* TrieMux: 500
* ChiRouter: 1500

Features:
* TinyRouter: limited
* HttpRouter: limited
* GorillaMux: rich
* TrieMux: limited
* ChiRouter: rich
