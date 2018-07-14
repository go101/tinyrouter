package tinyrouter

/*
 A tiny Go http router supporting custom parameters in paths.
*/

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// A Params value encapsulates the parameters passed from request URL path.
// This method will never panic.
type Params struct {
	kvs []string
}

// Value returns the parameter value corresponds to key.
// This method will never panic.
func (p Params) Value(key string) string {
	if key != "" {
		for i := 0; i < len(p.kvs); i += 2 {
			if p.kvs[i] == key {
				return p.kvs[i+1]
			}
		}
	}
	return ""
}

// ValueByIndex returns the parameter value corresponds to index i.
func (p Params) ValueByIndex(i int) string {
	if i < p.Num() {
		return p.kvs[i+i+1]
	}
	return ""
}

// Num returns the number of parameters.
func (p Params) Num() int {
	return len(p.kvs) >> 1
}

// To avoid being overwritten bu outer code.
type paramsContextKeyType struct{}

var paramsContextKey paramsContextKeyType

func requestWithParams(req *http.Request, kvs []string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), paramsContextKey, Params{kvs}))
}

// ParamsFromRequest returns parameters passed from URL path.
func ParamsFromRequest(req *http.Request) Params {
	p, _ := req.Context().Value(paramsContextKey).(Params)
	return p
}

type segment struct {
	// For fixed segment, this is the text needed to be matched exactly.
	// For wildcard segment, this is the parameter name.
	token string

	// The previous/next segments in path.
	nextInRow *segment

	// The segment (at the same column) in the next path.
	// Only used in initialization phase.
	nextInCol *segment

	// The first segment (at the same column) with a larger token, but
	// with the same length. A non-nil startLarger can't be wildcard.
	startLarger *segment

	// The first segment (at the same column) with a longer token.
	// A startLonger may be equal to startWildcard.
	startLonger *segment

	// The first wildcard segment (at the same column).
	// If set.startWildcard == seg, then seg is wildcard.
	startWildcard *segment

	// Which path this segment belongs to.
	path *path

	// For debug only.
	rowIndex int64
}

func (seg *segment) wildcard() bool {
	return seg.startWildcard == seg
}

func (seg *segment) next() *segment {
	if seg == nil {
		return nil
	}
	return seg.nextInRow
}

func (seg *segment) row() int {
	if seg == nil {
		return -1
	}
	return int(seg.rowIndex)
}

type path struct {
	raw       string // unparsed pattern
	next      *path
	segments  []*segment
	numParams int // how many
	handle    func(http.ResponseWriter, *http.Request)
}

func (p *path) String() string {
	return p.raw
}

func (p *path) makeParams(tokens []string) (kvs []string) {
	if p.numParams > 0 {
		kvs = make([]string, p.numParams<<1)
		k := 0
		for i, seg := range p.segments {
			if seg.wildcard() {
				kvs[k] = seg.token
				kvs[k+1] = tokens[i]
				k = k + 2
			}
		}
	}
	return
}

func compareSegments(sa, sb *segment) int {
	if sa.wildcard() && sb.wildcard() {
		return 0
	}

	if sa.wildcard() {
		return 1
	} else if sb.wildcard() {
		return -1
	} else if len(sa.token) > len(sb.token) {
		return 1
	} else if len(sa.token) < len(sb.token) {
		return -1
	}

	return strings.Compare(sa.token, sb.token)
}

func comparePaths(x, y *path) int {
	if len(x.segments) != len(y.segments) {
		panic("only paths with the same number of segments can be compared")
	}

	a := x.segments
	b := y.segments[:len(a)]
	for col := 0; col < len(a); col++ {
		if r := compareSegments(a[col], b[col]); r != 0 {
			return r
		}
	}

	return 0
}

func parsePath(r Route) *path {
	if len(r.Pattern) == 0 && r.Pattern[0] != '/' {
		panic("a pattern shell start with a slash: " + r.Pattern)
	}

	path := &path{raw: r.Pattern, handle: r.HandleFunc}

	buildSegment := func(pattern string, segs []*segment) (seg *segment) {
		if strings.HasPrefix(pattern, ":") {
			for _, seg := range segs {
				if seg.wildcard() && seg.token != "" && seg.token == pattern[1:] {
					panic("duplicated parameter name [" + pattern[1:] + "] in " + r.Pattern)
				}
			}
			seg = &segment{path: path, token: pattern[1:]}
			seg.startWildcard = seg
			path.numParams++
		} else {
			seg = &segment{path: path, token: pattern}
		}
		if len(segs) > 0 {
			segs[len(segs)-1].nextInRow = seg
		}
		return
	}

	var segs []*segment
	for pattern := r.Pattern[1:]; ; {
		i := strings.IndexRune(pattern, '/')
		if i >= 0 {
			segs = append(segs, buildSegment(pattern[:i], segs))
			pattern = pattern[i+1:]
		} else {
			segs = append(segs, buildSegment(pattern, segs))
			break
		}
	}
	if len(segs) > maxSegmentsInPath {
		panic("too many segments in path: " + r.Pattern)
	}
	path.segments = segs

	return path
}

func buildSegmentRelations(startSeg, endSeg *segment) {
	if startSeg == nil {
		return
	}

	seg, lastSeg, shortStart, smallerStart := startSeg, startSeg, startSeg, startSeg

	updateStartLargers := func() {
		if seg == nil || seg.wildcard() || len(lastSeg.token) != len(seg.token) {
			return
		}
		for smaller := smallerStart; smaller != seg; smaller = smaller.nextInCol {
			if smaller.wildcard() {
				panic("smaller is wildcard")
			}
			if lastSeg.wildcard() {
				panic("lastSeg is wildcard")
			}
			smaller.startLarger = seg
		}
	}

	updateStartLongers := func() {
		for short := shortStart; short != seg; short = short.nextInCol {
			//if short.wildcard() {panic("short is wildcard")}
			short.startLonger = seg
		}
	}

	updateStartWildcards := func() {
		for fixed := startSeg; fixed != seg; fixed = fixed.nextInCol {
			if fixed.wildcard() {
				panic("fixed is wildcard")
			}
			fixed.startWildcard = seg
		}
	}

	for ; seg != endSeg; lastSeg, seg = seg, seg.nextInCol {
		if seg.wildcard() {
			updateStartLongers()
			updateStartWildcards()
			buildSegmentRelations(smallerStart.next(), seg.next())
			break
		}

		if len(seg.token) > len(shortStart.token) {
			updateStartLargers()
			updateStartLongers()
			shortStart, smallerStart = seg, seg
			continue
		}

		if compareSegments(seg, smallerStart) > 0 {
			updateStartLargers()
			buildSegmentRelations(smallerStart.next(), seg.next())
			smallerStart = seg
		}
	}

	// Come here for two reasons: wildcard or endSeg encountered.
	if seg == endSeg {
		buildSegmentRelations(smallerStart.next(), endSeg.next())
		return
	}

	if seg == nil {
		panic("seg is nil")
	}
	if !seg.wildcard() {
		panic("seg is not wildcard")
	}
	buildSegmentRelations(seg.next(), endSeg.next())
}

func findHandlePath(tokens []string, entrySeg *segment) *path {
	for token, seg := tokens[0], entrySeg; seg != entrySeg.startWildcard; {
		if len(seg.token) > len(token) {
			break
		}
		if len(seg.token) < len(token) {
			seg = seg.startLonger
			continue
		}

		for k := 0; k < len(token); {
			if seg.token[k] > token[k] {
				goto Wildcard
			}
			if seg.token[k] < token[k] {
				seg = seg.startLarger
				if seg == nil {
					goto Wildcard
				}
				continue
			}
			k++
		}

		if seg.nextInRow == nil {
			return seg.path
		}

		path := findHandlePath(tokens[1:], seg.nextInRow)
		if path != nil {
			return path
		}

		goto Wildcard
	}

Wildcard:
	entrySeg = entrySeg.startWildcard
	if entrySeg == nil {
		return nil
	}

	if entrySeg.nextInRow == nil {
		return entrySeg.path
	}

	return findHandlePath(tokens[1:], entrySeg.nextInRow)
}

// Hard limit for number of segments in path.
const maxSegmentsInPath = 16

type tinyRouter struct {
	// First by the number of tokens, then by method.
	// Used in initialization phase and in dumping.
	pathsByNumToken [maxSegmentsInPath]map[string][]*path

	// Used in serving phase. The entry segment is paths[0].segments[0].
	entryByNumToken [maxSegmentsInPath]map[string]*segment

	// To avoid power exhausting attacks in request path parsing.
	maxNumTokens int

	// The default one in the standard http package is used on nil.
	notFoundHandleFunc http.HandlerFunc
}

// A Config value specifies the properties of a tinyRouter.
type Config struct {
	// This routing table
	Routes []Route

	// Not found handler function.
	NotFoundHandleFunc http.HandlerFunc

	// todo:
	// Ignore tailing slash or not.
	// Explicit routes have higher priorities.
	//IgnoreTailingSlash bool

	// todo:
	// The prefix to indicate whether a token is parameterized.
	// Blank means ":".
	//ParameterPrefix string
}

// A Route value specifies a request method, path pattern and
// the corresponding http handler function.
type Route struct {
	Method, Pattern string
	HandleFunc      http.HandlerFunc
}

// New returns *tinyRouter value, which is also a http.Handler value.
func New(c *Config) *tinyRouter {
	tr := &tinyRouter{notFoundHandleFunc: c.NotFoundHandleFunc}

	for _, r := range c.Routes {
		if r.HandleFunc == nil {
			panic("HandleFunc of a Route can't be nil")
		}
		rpath := parsePath(r)
		if len(rpath.segments) > tr.maxNumTokens {
			tr.maxNumTokens = len(rpath.segments)
		}
		pathsByMethod := tr.pathsByNumToken[len(rpath.segments)-1]
		if pathsByMethod == nil {
			pathsByMethod = make(map[string][]*path, 4)
		}
		pathsByMethod[r.Method] = append(pathsByMethod[r.Method], rpath)
		tr.pathsByNumToken[len(rpath.segments)-1] = pathsByMethod
	}

	for numSegments, pathsByMethod := range tr.pathsByNumToken[:] {
		if len(pathsByMethod) == 0 {
			continue
		}

		entryByNumMethod := make(map[string]*segment, len(pathsByMethod))
		for method, paths := range pathsByMethod {
			sort.Slice(paths, func(i, j int) bool {
				return comparePaths(paths[i], paths[j]) < 0
			})

			for prevPath, i, row := paths[0], 1, int64(0); i < len(paths); i++ {
				path := paths[i]
				if comparePaths(prevPath, path) == 0 {
					panic(fmt.Sprintf("Equal paths are not allowed:\n   %s\n   %s", prevPath, path))
				}

				prevSeg, seg := prevPath.segments[0], path.segments[0]
				for seg != nil {
					prevSeg.rowIndex, seg.rowIndex = row, row+1 // for debug
					prevSeg.nextInCol = seg
					prevSeg, seg = prevSeg.nextInRow, seg.nextInRow
				}

				prevPath.next = path
				prevPath = path
				row++
			}

			entrySegment := paths[0].segments[0]
			entryByNumMethod[method] = entrySegment
			tr.entryByNumToken[numSegments] = entryByNumMethod

			buildSegmentRelations(entrySegment, nil)
		}
	}

	return tr
}

// Dump is for debug purpose.
func (tr *tinyRouter) Dump() string {
	var b strings.Builder
	for i, pathsByMethod := range tr.pathsByNumToken[:] {
		for method, paths := range pathsByMethod {
			if len(paths) == 0 {
				continue
			}

			b.WriteString(fmt.Sprintf("\nmethod %s with %d tokens:", method, i+1))
			for i, path := range paths {
				b.WriteString(fmt.Sprint("\n   ", i, "> "))
				for _, seg := range path.segments {
					b.WriteString("[")
					if seg.wildcard() {
						b.WriteString(":")
					}
					b.WriteString(seg.token)
					b.WriteString(" ")
					b.WriteString(strconv.Itoa(seg.startLarger.row()))
					b.WriteString(" ")
					b.WriteString(strconv.Itoa(seg.startLonger.row()))
					b.WriteString(" ")
					b.WriteString(strconv.Itoa(seg.startWildcard.row()))
					b.WriteString("]")
				}
			}
		}
	}
	return b.String()
}

// ServeHTTP lets *tinyRouter implement http.Handler interface.
func (tr *tinyRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	urlPath := req.URL.Path[1:]
	if len(urlPath) > 1024 {
		urlPath = urlPath[:1024]
	}

	tokens := strings.SplitN(urlPath, "/", tr.maxNumTokens)
	entryByMethod := tr.entryByNumToken[len(tokens)-1]
	if len(entryByMethod) == 0 {
		w.Write([]byte("no routes for path with " + strconv.Itoa(len(tokens)) + " tokens\n"))
		return
	}

	entrySegment := entryByMethod[req.Method]
	if entryByMethod == nil {
		w.Write([]byte("no routes for method: " + req.Method + " with " + strconv.Itoa(len(tokens)) + " tokens\n"))
		return
	}

	path := findHandlePath(tokens, entrySegment)
	if path == nil {
		if tr.notFoundHandleFunc == nil {
			http.NotFound(w, req)
		} else {
			tr.notFoundHandleFunc(w, req)
		}
		return
	}

	if path.numParams > 0 {
		req = requestWithParams(req, path.makeParams(tokens))
	}
	path.handle(w, req)
}
