/**
 Copyright 2016 Emmanuel Keller

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package main

import (
	"log"
	"net/http"
	"net/url"
	"github.com/gorilla/mux"
	"encoding/json"
	"strconv"
	"fmt"
	"io/ioutil"
	"time"
)

var languages = []string{"C++", "Java", "Go", "Rust", "Swift"}

func CountProjects(language string, year int, month int) (int, error) {

	// The public GH API rate is 10r eq/minutes
	time.Sleep(7 * time.Second)

	const GhEndPoint = "https://api.github.com/search/repositories?q=language:"

	u := GhEndPoint + url.QueryEscape(language) + "+pushed:" + fmt.Sprintf("%04d-%02d", year, month)
	resp, err := http.Get(u)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if (resp.StatusCode != 200) {
		return 0, fmt.Errorf("Wrong GH API status code: %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	type GHSearchResult struct {
		TotalCount int `json:"total_count"`
	}

	var result GHSearchResult
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return 0, err
	}

	return result.TotalCount, nil
}

func MainHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		http.Error(w, vars["year"] + " is not a numeric value", http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(http.StatusOK)

	for _, language := range languages {
		fmt.Fprint(w, language)
		for month := 1; month <= 12; month++ {
			count, err := CountProjects(language, year, month)
			if err != nil {
				http.Error(w, "Error on language " + language + " " + err.Error(), http.StatusNotAcceptable)
				return
			}
			log.Printf("%04d-%02d\t%s\t%d", year, month, language, count)
			fmt.Fprintf(w, "\t%d", count)
		}
		fmt.Fprintln(w)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{year}", MainHandler)
	log.Fatal(http.ListenAndServe(":8000", r))
}