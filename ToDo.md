

* change `map[string][]*path` and `map[string]*segment` to
`[NumMethods][]*path` and `[NumMethods]*segment`.
* use requestToken.lastSmaller and requestToken.cmpCursor to optimize.
* add examples in docs
