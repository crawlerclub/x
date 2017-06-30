package store

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

func buildIndexMapping() *mapping.IndexMappingImpl {
	indexMapping := bleve.NewIndexMapping()

	crawlerMapping := bleve.NewDocumentMapping()

	parseConfsMapping := bleve.NewDocumentMapping()
	parseConfsMapping.Enabled = false
	parseConfsMapping.Dynamic = false
	crawlerMapping.AddSubDocumentMapping("parse_confs", parseConfsMapping)

	indexMapping.AddDocumentMapping("crawler", crawlerMapping)

	taskMapping := bleve.NewDocumentMapping()

	indexMapping.AddDocumentMapping("task", taskMapping)

	return indexMapping
}
