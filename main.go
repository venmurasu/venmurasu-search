//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	_ "expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	bleveHttp "github.com/blevesearch/bleve/http"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

var batchSize = flag.Int("batchSize", 150, "batch size for indexing")
var bindAddr = flag.String("addr", ":8094", "http listen address")
var jsonDir = flag.String("jsonDir", "/Users/mahendra/products/venmurasu/bleve_data", "json directory")
var indexPath = flag.String("index", "vensearch.bleve", "index path")
var staticEtag = flag.String("staticEtag", "", "A static etag value.")
var staticPath = flag.String("static", "static/", "Path to the static content")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write mem profile to file")

func main() {

	flag.Parse()

	log.Printf("GOMAXPROCS: %d", runtime.GOMAXPROCS(-1))

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		//w.Write([]byte("hi"))
		http.Redirect(w, r, "/static/index.html", http.StatusMovedPermanently)
	})
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	//open the index
	venmurasuIndex, err := bleve.Open(*indexPath)

	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index...")
		// create a mapping
		indexMapping, err := buildIndexMapping()
		if err != nil {
			log.Fatal(err)
		}
		//venmurasuIndex, err = bleve.New(*indexPath, indexMapping)
		venmurasuIndex, err := bleve.NewUsing(*indexPath, indexMapping, scorch.Name, scorch.Name, map[string]interface{}{
			"forceSegmentType":    "zap",
			"forceSegmentVersion": 12,
		})

		if err != nil {
			log.Fatal(err)
		}

		// index data in the background
		go func() {
			err = indexVenmurasu(venmurasuIndex)
			if err != nil {
				log.Fatal(err)
			}
			pprof.StopCPUProfile()
			if *memprofile != "" {
				f, err := os.Create(*memprofile)
				if err != nil {
					log.Fatal(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
			}
		}()
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Opening existing index...")
	}

	// create a router to serve static files

	// workDir, _ := os.Getwd()
	// filesDir := http.Dir(filepath.Join(workDir, "static"))

	FileServer(r, "/static", "./static/")
	//FileServer(r, "/static", "./static/")
	//FileServer(r, "/", "/static/")

	// add the API
	bleveHttp.RegisterIndexName("venmurasu", venmurasuIndex)
	searchHandler := bleveHttp.NewSearchHandler("venmurasu")

	// router.Handle("/api/search", searchHandler).Methods("POST")

	r.Route("/api", func(r chi.Router) {
		r.With(searchParams).Post("/search", searchHandler.ServeHTTP) // POST /articles
		r.Post("/stdsearch", searchHandler.ServeHTTP)
	})

	// start the HTTP server
	// http.Handle("/", router)
	// log.Printf("Listening on %v", *bindAddr)
	// log.Fatal(http.ListenAndServe(*bindAddr, nil))

	log.Printf("Listening on %s", ":8094")
	http.ListenAndServe(":8094", r)
}

// func Search(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("inside search")
// 	venmurasuIndex, err := bleve.Open(*indexPath)
// 	fmt.Println("after open")
// 	if err != nil {
// 		fmt.Println("err opening index==>", err)
// 		panic(err)
// 	}

// 	query := bleve.NewMatchQuery("இளைய யாதவர்")
// 	searchRequest := bleve.NewSearchRequest(query)
// 	searchResults, err := venmurasuIndex.Search(searchRequest)
// 	if err != nil {
// 		fmt.Println("errerrerrerr==>", err)
// 		panic(err)
// 	}

// 	fmt.Println(searchResults)

// }

type ReqJSON struct {
	SearchQry string `json:"search"`
	From      int    `json:"from"`
	Size      int    `json:"size"`
}

func searchParams(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var from, size int
		var source, tag, search, bookno, bookname string
		var searchQry []string
		var req ReqJSON
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)

		if err != nil {
			fmt.Printf("error reading request body in middleware: %v", err)
			return
		}

		source = req.SearchQry
		from = req.From
		size = req.Size
		if size == 0 {
			size = 10
		}
		fmt.Println("sourcesourcesource==>", source)

		if strings.Contains(source, "tags:") {
			tagsre := regexp.MustCompile(`(?P<tags>tags:(.*\b?))`)
			if len(tagsre.FindStringSubmatch(source)) > 0 {
				tag = strings.Replace(tagsre.FindStringSubmatch(source)[0], "tags:", "", 1)
			}
			searchQry = append(searchQry, `{
				"term": "`+strings.TrimRight(tag, " ")+`",
				"field": "tags",
				"boost": 1
			}`)

		}
		if strings.Contains(source, "search:") || !strings.Contains(source, ":") {
			searchre := regexp.MustCompile(`(?P<search>search:'(.*?)')`)

			if strings.Contains(source, "search:") {
				search = strings.Replace(strings.Replace(searchre.FindStringSubmatch(source)[0], "search:", "", 1), "'", "", 2)
			} else {
				search = strings.Replace(strings.Replace(source, "'", "", 2), "\"", "", 2)
			}
			searchQry = append(searchQry, `{                
                "match_phrase": "`+search+`"  ,
				"field": "_all",
				"boost":1
            }`)
		}
		if strings.Contains(source, "bookno:") {
			booknore := regexp.MustCompile(`(?P<bookno>bookno:(.*\b?))`)
			if len(booknore.FindStringSubmatch(source)) > 0 {
				bookno = strings.Replace(booknore.FindStringSubmatch(source)[0], "bookno:", "", 1)
			}
			searchQry = append(searchQry, `{
				"term": "`+strings.TrimRight(bookno, " ")+`",
				"field": "bookno",
				"boost": 1
			}`)
		}
		if strings.Contains(source, "bookname:") {
			booknamere := regexp.MustCompile(`(?P<bookname>bookname:(.*\b?))`)
			if len(booknamere.FindStringSubmatch(source)) > 0 {
				bookname = strings.Replace(booknamere.FindStringSubmatch(source)[0], "bookname:", "", 1)
			}
			searchQry = append(searchQry, `{
				"term": "`+strings.TrimRight(bookname, " ")+`",
				"field": "bookname",
				"boost": 1
			}`)

		}
		fmt.Println(tag, search, bookno, bookname)
		searchQryFinal := strings.Join(searchQry, ",")

		qry :=
			fmt.Sprintf(`{
			"from": %d,
			"explain": true,
			"size": %d,	
			"query": {
			  "must": {
				"conjuncts": [   
				%s
				] 
			  } 
			},
			"highlight": {},
			"fields": ["bookno", "chapter", "bookname", "sectionno", "sectionname", "published_on" , "book"]

		  }`, from, size, searchQryFinal)
		fmt.Println("Search query===>", qry)

		formattedQry := []byte(qry)

		r.Body = ioutil.NopCloser(bytes.NewReader(formattedQry))

		// Here we are pssing our custom response writer to the next http handler.

		// _, err := ioutil.ReadAll(r.Body)
		// if err != nil {
		// 	fmt.Printf("error reading request body: %v", err)
		// 	return
		// }

		// r.Body = `{
		// 	"query": {
		// 	  "must": {
		// 		"conjuncts": [

		// 				  {
		// 			"match": "ஜனமேஜயன்",
		// 			"field": "_all",
		// 			"boost": 1
		// 		  }

		// 		],
		// 		"boost": 1
		// 	  },
		// 	  "boost": 1
		// 	},
		// 	"highlight": {},
		// 	"size": 10
		//   }`
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}

func indexVenmurasu(i bleve.Index) error {

	// open the directory
	dirEntries, err := ioutil.ReadDir(*jsonDir)
	if err != nil {
		return err
	}

	// walk the directory entries for indexing
	log.Printf("Indexing...")
	count := 0
	startTime := time.Now()
	batch := i.NewBatch()
	batchCount := 0
	for _, dirEntry := range dirEntries {
		filename := dirEntry.Name()

		// read the bytes
		jsonBytes, err := ioutil.ReadFile(*jsonDir + "/" + filename)
		if err != nil {
			return err
		}
		// parse bytes as json
		var jsonDoc interface{}
		err = json.Unmarshal(jsonBytes, &jsonDoc)
		if err != nil {
			return err
		}

		ext := filepath.Ext(filename)
		docID := filename[:(len(filename) - len(ext))]
		fmt.Println("Indexing docid==>", docID)
		batch.Index(docID, jsonDoc)
		batchCount++

		if batchCount >= *batchSize {
			err = i.Batch(batch)
			if err != nil {
				return err
			}
			batch = i.NewBatch()
			batchCount = 0
		}
		count++
		if count%10 == 0 {
			indexDuration := time.Since(startTime)
			indexDurationSeconds := float64(indexDuration) / float64(time.Second)
			timePerDoc := float64(indexDuration) / float64(count)
			log.Printf("Indexed %d documents, in %.2fs (average %.2fms/doc)", count, indexDurationSeconds, timePerDoc/float64(time.Millisecond))
		}
	}
	// flush the last batch
	if batchCount > 0 {
		err = i.Batch(batch)
		if err != nil {
			log.Fatal(err)
		}
	}
	indexDuration := time.Since(startTime)
	indexDurationSeconds := float64(indexDuration) / float64(time.Second)
	timePerDoc := float64(indexDuration) / float64(count)
	log.Printf("Indexed %d documents, in %.2fs (average %.2fms/doc)", count, indexDurationSeconds, timePerDoc/float64(time.Millisecond))
	return nil
}

func FileServer(r chi.Router, public string, static string) {

	if strings.ContainsAny(public, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	root, _ := filepath.Abs(static)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		panic("Static Documents Directory Not Found")
	}

	fs := http.StripPrefix(public, http.FileServer(http.Dir(root)))

	if public != "/" && public[len(public)-1] != '/' {
		r.Get(public, http.RedirectHandler(public+"/", 301).ServeHTTP)
		public += "/"
	}

	r.Get(public+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := strings.Replace(r.RequestURI, public, "/", 1)
		if _, err := os.Stat(root + file); os.IsNotExist(err) {
			http.ServeFile(w, r, path.Join(root, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}))
}
