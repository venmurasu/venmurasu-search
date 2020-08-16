package main

import (
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/ta"
	"github.com/blevesearch/bleve/analysis/tokenizer/unicode"
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
	venMapping.AddFieldMappingsAt("book", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("sectionno", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("sectionname", keywordFieldMapping)
	venMapping.AddFieldMappingsAt("published_on", datetimeFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("venmurasu", venMapping)

	var err error
	err = indexMapping.AddCustomAnalyzer("customta",
		map[string]interface{}{
			"type":      custom.Name,
			"tokenizer": unicode.Name,
			"token_filters": []interface{}{
				ta.StopName,
				ta.SnowballStemmerName,
				`normalize_in`,
			},
		})
	if err != nil {
		return nil, err
	}

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = "customta"
	fmt.Printf("IndexMappingImpl===>%+v", indexMapping)
	return indexMapping, nil
}
