package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	defaultURL = `http://localhost:3030/test/query`
)

func main() {
	query := flag.String("q", "", "search query")
	url := flag.String("api", defaultURL, "api URL")
	user := flag.String("user", "000000", "user attribute bitmap")

	var queryToRun string
	if *query == "" {
		queryToRun = fmt.Sprintf("%s%s%s", `SELECT DISTINCT ?g WHERE { GRAPH ?g {?s ?p ?o .} FILTER( REGEX(STR(?g), REPLACE("`, *user, `", "1", "\\.")+"$")) }`)
	} else {
		queryToRun = *query
	}
	queryToRun = `PREFIX foo: <http://localhost:3030/test/> SELECT DISTINCT ?s FROM foo:000011 WHERE { ?s ?p ?o }`
	r, duration := runQueryWithTimer(queryToRun, *url)
	if r != nil {
		defer r.Body.Close()
		sc := bufio.NewScanner(r.Body)
		i := 0
		for sc.Scan() && i < 10 {
			fmt.Println(sc.Text())
			// i++
		}
	} else {
		fmt.Println("No response body")
	}
	fmt.Println(duration)
}

func runQueryWithTimer(query, url string) (*http.Response, time.Duration) {
	start := time.Now()
	r := runQuery(query, url)
	return r, time.Since(start)
}

func runQuery(query, url string) *http.Response {
	client := http.DefaultClient
	r := strings.NewReader(query)
	resp, err := client.Post(url, "application/sparql-query", r)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return resp
}
