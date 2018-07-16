

* change `map[string][]*path` and `map[string]*segment` to
`[NumMethods][]*path` and `[NumMethods]*segment`.
* use a new url path token type to avoid one allocation in creating params.
* add a path.wildcardSegments to optimize Params.Value() method.
