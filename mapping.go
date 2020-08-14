package main

import (
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/ta"
	"github.com/blevesearch/bleve/mapping"
)

// "bookno": 1,
// "chapter": 1,
// "bookname": "முதற்கனல்",
// "sectionno": 1,
// "sectionname": "பகுதி ஒன்று : வேள்விமுகம்",
// "published_on": "01-01-2014",

func buildIndexMapping() (mapping.IndexMapping, error) {

	// a generic reusable mapping for tamil text
	tamilTextFieldMapping := bleve.NewTextFieldMapping()
	tamilTextFieldMapping.Analyzer = ta.AnalyzerName

	// a generic reusable mapping for keyword text
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	datetimeFieldMapping := bleve.NewDateTimeFieldMapping()

	venMapping := bleve.NewDocumentMapping()

	venMapping.AddFieldMappingsAt("bookno", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("chapter", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("bookname", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("sectionno", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("sectionname", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("published_on", datetimeFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("beer", venMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = "ta"
	fmt.Printf("IndexMappingImpl===>%+v", indexMapping)
	return indexMapping, nil
}
