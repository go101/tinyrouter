



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
the HttpRouter is the best choice.

### Other Comparisons

Use standard `http.HandlerFunc` as handler functions?
* HttpRouter: No
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
