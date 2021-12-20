package main

import (
	"encoding/json"
	_ "expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/index/scorch"
)

// generate the bleve index for the given json.
func main() {
	var batchSize = flag.Int("batchSize", 150, "batch size for indexing")
	var jsonDir = flag.String("jsonDir", "data/venmurasu-json", "json directory")
	var indexPath = flag.String("index", "vensearch.bleve", "index path")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write mem profile to file")

	flag.Parse()

	log.Printf("GOMAXPROCS: %d", runtime.GOMAXPROCS(-1))

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	//generate the index
	// venmurasuIndex, err := bleve.Open(*indexPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Printf("Creating new index... %s", *indexPath)
	// create a mapping
	indexMapping, err := buildIndexMapping()
	if err != nil {
		log.Fatal(err)
	}

	venmurasuIndex, err := bleve.NewUsing(*indexPath, indexMapping, scorch.Name, scorch.Name, map[string]interface{}{
		"forceSegmentType":    "zap",
		"forceSegmentVersion": 12,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = indexVenmurasu(venmurasuIndex, *jsonDir, *batchSize)
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

}

func indexVenmurasu(i bleve.Index, jsonDir string, batchSize int) error {

	// open the directory
	dirEntries, err := ioutil.ReadDir(jsonDir)
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
		jsonBytes, err := ioutil.ReadFile(jsonDir + "/" + filename)
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

		if batchCount >= batchSize {
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