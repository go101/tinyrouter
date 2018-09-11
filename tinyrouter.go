// Tinyrouter is Go http router supporting custom parameters in paths.
// The implementation contains only 500 lines of code.
package tinyrouter

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// A Params encapsulates the parameters in request URL path.
type Params struct {
	path   *path
	tokens []string
}

// Value returns the parameter value corresponds to key.
// This method will never panic.
func (p Params) Value(key string) string {
	if p.path != nil {
		for _, seg := range p.path.wildcards {
			if seg.token == key {
				return p.tokens[seg.colIndex]
			}
		}
	}
	return ""
}

// ValueByIndex returns the parameter value corresponds to index i.
// This method will never panic.
func (p Params) ValueByIndex(i int) string {
	if p.path != nil && i >= 0 && i < len(p.path.wildcards) {
		return p.tokens[p.path.wildcards[i].colIndex]
	}
	return ""
}

// Convert a Params to a map[string]string and a []string.
// Mainly for debug purpose.
func (p Params) ToMapAndSlice() (kvs map[string]string, vs []string) {
	if p.path != nil {
		kvs = make(map[string]string, p.path.numParams)
		vs = make([]string, 0, p.path.numParams)
		for _, seg := range p.path.segments {
			if seg.wildcard() {
				vs = append(vs, p.tokens[seg.colIndex])
				kvs[seg.token] = p.tokens[seg.colIndex]
			}
		}
	}
	return
}

// To avoid being overwritten by outer code.
type paramsKeyType struct{}

// PathParams returns parameters passed from URL path.
func PathParams(req *http.Request) Params {
	p, _ := req.Context().Value(paramsKeyType{}).(Params)
	return p
}

type segment struct {
	// Which path this segment belongs to.
	path *path

	// For fixed segment, this is the text needed to be matched exactly.
	// For wildcard segment, this is the parameter name.
	token string

	// The next segment in path and the same0column segment in the next paths.
	nextInRow, nextInCol *segment

	// The first segment (at the same column) with a larger token, but
	// with the same length. A non-nil startLarger can't be wildcard.
	startLarger *segment

	// The first segment (at the same column) with a longer token.
	// A startLonger may be equal to startWildcard.
	startLonger *segment

	// The first wildcard segment (at the same column).
	// If seg.startWildcard == seg, then segment seg is wildcard.
	startWildcard *segment

	// rowIndex is for debug only.
	rowIndex, colIndex int32

	// How many equal prefix bytes with startLarger.
	numSameBytes int32
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
	raw       string     // unparsed pattern
	segments  []*segment // []segment is better? Need benchmark. (or [][]segments for a path group?)
	wildcards []*segment // for fast parameter value look-up
	handle    func(http.ResponseWriter, *http.Request)
	numParams int32 // how many wildcard segments in this path
	row       int32 // row index in a path group
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
	if len(r.Pattern) == 0 || r.Pattern[0] != '/' {
		panic("a pattern shell start with a slash: " + r.Pattern)
	}

	path := &path{raw: r.Pattern, handle: r.HandleFunc}

	buildSegment := func(pattern string, segs []*segment) (seg *segment) {
		if strings.HasPrefix(pattern, ":") {
			for _, seg := range segs {
				if seg.wildcard() && seg.token == pattern[1:] {
					panic("duplicated parameter name [" + pattern[1:] + "] in " + r.Pattern)
				}
			}
			seg = &segment{path: path, token: pattern[1:]}
			seg.startWildcard = seg
			path.numParams++
			path.wildcards = append(path.wildcards, seg)
		} else {
			seg = &segment{path: path, token: pattern}
		}

		if seg.colIndex = int32(len(segs)); seg.colIndex > 0 {
			segs[seg.colIndex-1].nextInRow = seg
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
			buildSegmentRelations(smallerStart.next(), seg.next())
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

		var k, n = uint(0), uint(len(token))
	Next:
		if len(seg.token) == len(token) { // BCE
			for k < n {
				if seg.token[k] > token[k] { // BCEed
					goto Wildcard
				}
				if seg.token[k] < token[k] {
					if seg.startLarger == nil || seg.numSameBytes < int32(k) {
						goto Wildcard
					}
					seg = seg.startLarger
					goto Next
				}
				k++
			}
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

// Hard limit for maximum number of segments in path.
const maxSegmentsInPath = 32

type TinyRouter struct {
	// First by the number of tokens, then by method.
	// Used in initialization phase and in dumping.
	pathsByMethod map[string]*[maxSegmentsInPath][]*path

	// Used in serving phase. The entry segment is paths[0].segments[0].
	entryByMethod map[string]*[maxSegmentsInPath]*segment

	// To avoid power exhausting attacks in request path parsing.
	maxNumTokens int

	// The default one in the standard http package is used on nil.
	othersHandleFunc http.HandlerFunc
}

// A Config value specifies the properties of a TinyRouter.
type Config struct {
	// This routing table
	Routes []Route

	// Handler function for unmatched paths.
	// Nil means http.NotFound.
	OthersHandleFunc http.HandlerFunc

	// todo:
	// Ignore tailing slash or not.
	// Explicit routes have higher priorities.
	//IgnoreTailingSlash bool
}

// A Route value specifies a request method, path pattern and
// the corresponding http handler function.
type Route struct {
	Method, Pattern string
	HandleFunc      http.HandlerFunc
}

// New returns a *TinyRouter value, which is also a http.Handler value.
func New(c Config) *TinyRouter {
	tr := &TinyRouter{othersHandleFunc: c.OthersHandleFunc}
	if tr.othersHandleFunc == nil {
		tr.othersHandleFunc = http.NotFound
	}
	tr.pathsByMethod = make(map[string]*[maxSegmentsInPath][]*path, 8)
	tr.entryByMethod = make(map[string]*[maxSegmentsInPath]*segment, 8)

	for _, r := range c.Routes {
		if r.HandleFunc == nil {
			panic("HandleFunc of a Route can't be nil")
		}
		rpath := parsePath(r)
		if len(rpath.segments) > tr.maxNumTokens {
			tr.maxNumTokens = len(rpath.segments)
		}
		if tr.entryByMethod[r.Method] == nil {
			tr.pathsByMethod[r.Method] = &[maxSegmentsInPath][]*path{}
			tr.entryByMethod[r.Method] = &[maxSegmentsInPath]*segment{}
		}
		paths := tr.pathsByMethod[r.Method][len(rpath.segments)-1]
		if paths == nil {
			paths = make([]*path, 0, 4)
		}
		tr.pathsByMethod[r.Method][len(rpath.segments)-1] = append(paths, rpath)
	}

	for method, pathsByNumTokens := range tr.pathsByMethod {
		for numTokens, paths := range pathsByNumTokens {
			if paths == nil {
				continue
			}

			sort.Slice(paths, func(i, j int) bool {
				return comparePaths(paths[i], paths[j]) < 0
			})

			for prevPath, i, row := paths[0], 1, int32(0); i < len(paths); i++ {
				path := paths[i]
				if comparePaths(prevPath, path) == 0 {
					panic(fmt.Sprintf("Equal paths are not allowed:\n   %s\n   %s", prevPath.raw, path.raw))
				}

				prevSeg, seg := prevPath.segments[0], path.segments[0]
				for seg != nil {
					prevSeg.rowIndex, seg.rowIndex = row, row+1 // for debug
					prevSeg.nextInCol = seg
					prevSeg, seg = prevSeg.nextInRow, seg.nextInRow
				}

				prevPath, row = path, row+1
			}
			tr.entryByMethod[method][numTokens] = paths[0].segments[0]

			buildSegmentRelations(paths[0].segments[0], nil)

			statSamePrefixBytes := func(a, b string, num *int32) {
				for ; *num < int32(len(a)) && a[*num] == b[*num]; *num++ {
				}
			}
			for col := 0; col < len(paths[0].segments); col++ {
				for row := 0; row < len(paths); row++ {
					seg := paths[row].segments[col]
					if seg.startLarger != nil {
						statSamePrefixBytes(seg.token, seg.startLarger.token, &seg.numSameBytes)
					}
				}
			}
		}
	}
	return tr
}

// DumpInfo is for debug purpose.
func (tr *TinyRouter) DumpInfo() string {
	var b strings.Builder
	for method, pathsByNumTokens := range tr.pathsByMethod {
		for numTokens, paths := range pathsByNumTokens {
			if len(paths) == 0 {
				continue
			}

			b.WriteString(fmt.Sprintf("\nmethod %s with %d tokens:", method, numTokens+1))
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
					b.WriteString(" ")
					b.WriteString(strconv.Itoa(int(seg.numSameBytes)))
					b.WriteString("]")
				}
			}
		}
	}
	return b.String()
}

// ServeHTTP lets *TinyRouter implement http.Handler interface.
func (tr *TinyRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	urlPath := req.URL.Path[1:]
	if len(urlPath) > 1024 {
		urlPath = urlPath[:1024]
	}

	entryByNumTokens := tr.entryByMethod[req.Method]
	if entryByNumTokens == nil {
		tr.othersHandleFunc(w, req)
		return
	}

	tokens := strings.SplitN(urlPath, "/", tr.maxNumTokens)
	entrySegment := entryByNumTokens[len(tokens)-1]
	if entrySegment == nil {
		tr.othersHandleFunc(w, req)
		return
	}

	path := findHandlePath(tokens, entrySegment)
	if path == nil {
		tr.othersHandleFunc(w, req)
		return
	}

	if path.numParams > 0 {
		req = req.WithContext(context.WithValue(req.Context(), paramsKeyType{}, Params{path, tokens}))
	}
	path.handle(w, req)
}
