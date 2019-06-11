package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

var (
	filename = `/home/james.ritchie/Downloads/Fuseki/data.nt`
	graphs   = 255
	url      = `http://localhost:3030/test/query`
	userIDs  []string
)

func TestFuseki(t *testing.T) {
	t.Run("Load Data", func(*testing.T) {
		if run(filename, graphs, true) != nil {
			t.FailNow()
		}
	})

	makeUserIDs()

	t.Run("Base", FusekiBaselineQueries)
	t.Run("Regex", FusekiRegexQueries)
	t.Run("FROM", FusekiFROMQueries)
}

func FusekiBaselineQueries(t *testing.T) {
	zeroAttributes := strings.Repeat("0", int(maxNameLength()))
	testQueries := []string{
		`SELECT DISTINCT ?p WHERE { ?s ?p ?o }`,
		`SELECT DISTINCT ?p WHERE { GRAPH ?g { ?s ?p ?o } }`,
		`PREFIX foo: <http://localhost:3030/test/> SELECT DISTINCT ?p FROM foo:` + zeroAttributes + ` WHERE { ?s ?p ?o}`,
	}
	runQueries(t, testQueries, url)
}

func FusekiRegexQueries(t *testing.T) {
	start := `SELECT DISTINCT ?p WHERE { GRAPH ?g {?s ?p ?o .} FILTER( REGEX(STR(?g), REPLACE("`
	end := `", "1", "\\.")+"$")) }`
	testQueries := buildQueries(start, end, userIDs)
	runQueries(t, testQueries, url)
}

func FusekiFROMQueries(t *testing.T) {
	start := `PREFIX foo: <http://localhost:3030/test/> SELECT DISTINCT ?p `
	end := ` WHERE { ?s ?p ?o .}`
	var inners, testQueries []string
	inners = buildFROMClauses()
	testQueries = buildQueries(start, end, inners)
	runQueries(t, testQueries, url)
}

func runQueries(t *testing.T, queries []string, url string) {
	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			r, _ := runQueryWithTimer(query, url)
			if r == nil {
				t.Fail()
			}
		})
	}
}

func makeUserIDs() {
	highestNameLength := maxNameLength()
	userIDs = make([]string, highestNameLength+1)
	for i := int64(0); i <= highestNameLength; i++ {
		val := int64(math.Pow(2, float64(i)) - 1)
		userIDs[i] = asBinary(val, highestNameLength)
	}
}

func maxNameLength() int64 {
	return int64(len(strconv.FormatInt(int64(graphs), 2)))
}

func buildQueries(start, end string, inners []string) []string {
	queries := make([]string, len(inners))
	for i, inner := range inners {
		queries[i] = start + inner + end
	}
	return queries
}

func buildFROMClauses() []string {
	clauses := make([]string, len(userIDs))
	for i, bitmap := range userIDs {
		var clause string
		graphs := getVisibleGraphs(bitmap)
		for _, graph := range graphs {
			clause = fmt.Sprintf("%s FROM foo:%s", clause, graph)
		}
		clauses[i] = clause
	}
	return clauses
}

func asBinary(value, length int64) string {
	bin := strconv.FormatInt(int64(value), 2)
	zeros := strings.Repeat("0", int(math.Max(0, float64(int(length)-len(bin)))))
	return fmt.Sprintf("%s%s", zeros, bin)
}

func getVisibleGraphs(userAttribute string) []string {
	binCount := len(strings.ReplaceAll(userAttribute, "0", ""))
	// results := []string{strings.Repeat("0", len(userAttribute))}
	var results []string
	base := strings.ReplaceAll(userAttribute, "1", "%s")
	for i := 1; i < int(math.Pow(2, float64(binCount))); i++ {
		iAsBitmap := bitMapOfLength(i, binCount)
		var bits []interface{}
		for _, c := range iAsBitmap {
			bits = append(bits, string(c))
		}
		match := fmt.Sprintf(base, bits...)
		results = append(results, match)
	}
	return results
}

func bitMap(value int) string {
	return strconv.FormatInt(int64(value), 2)
}

func bitMapOfLength(value, length int) string {
	bitmap := bitMap(value)
	if len(bitmap) < length {
		zeros := strings.Repeat("0", length-len(bitmap))
		bitmap = zeros + bitmap
	}
	return bitmap
}
