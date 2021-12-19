module github.com/cmrajan/venmurasu-search

go 1.14

require (
	github.com/blevesearch/bleve v1.0.14
	github.com/blevesearch/bleve/v2 v2.3.0
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/render v1.0.1
)

replace github.com/blevesearch/bleve/v2 v2.3.0 => github.com/arulrajnet/bleve/v2 v2.3.1-0.20211219105224-86317d19abde
