package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
)

const (
	elasticIndexName = "elecandgas"
)

type DocumentResponse struct {
	CreatedAt            time.Time `json:"created_at"`
	TRANSACTION_TYPE     string    `json:"TRANSACTION_TYPE"`
	FILE_TYPE            string    `json:"FILE_TYPE"`
	FILE_DATE            string    `json:"FILE_DATE"`
	FILE_NUM             string    `json:"FILE_NUM"`
	MPO_REFERENCE        string    `json:"MPO_REFERENCE"`
	SERIAL_NUMBER        string    `json:"SERIAL_NUMBER"`
	SUB_BUILDING         string    `json:"SUB_BUILDING"`
	BUILDING_NAME        string    `json:"BUILDING_NAME"`
	DELIVER_POINT_ALIAS  string    `json:"DELIVER_POINT_ALIAS"`
	BUILDING_NUMBER      string    `json:"BUILDING_NUMBER"`
	DEPENDENT_STREET     string    `json:"DEPENDENT_STREET"`
	PRINCIPAL_STREET     string    `json:"PRINCIPAL_STREET"`
	DBL_DPNDT_LOCLTY     string    `json:"DBL_DPNDT_LOCLTY"`
	DEPENDENT_LOCALITY   string    `json:"DEPENDENT_LOCALITY"`
	POST_TOWN            string    `json:"POST_TOWN"`
	COUNTY               string    `json:"COUNTY"`
	OUTCODE              string    `json:"OUTCODE"`
	INCODE               string    `json:"INCODE"`
	LARGE_SITE_INDICATOR string    `json:"LARGE_SITE_INDICATOR"`
	IGT                  string    `json:"IGT"`
}

type SearchResponse struct {
	Time      string             `json:"time"`
	Hits      string             `json:"hits"`
	Documents []DocumentResponse `json:"documents"`
}

var (
	elasticClient *elastic.Client
)

func main() {
	var err error
	// Create Elastic client and wait for Elasticsearch to be ready
	for {
		elasticClient, err = elastic.NewClient(
			elastic.SetURL("https://search-ampoweruk-search-ndtvt3itb3fqy7ais3vu4v53me.eu-west-1.es.amazonaws.com"),
			elastic.SetSniff(false),
		)
		if err != nil {
			log.Println(err)
			// Retry every 3 seconds
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	// Start HTTP server
	r := gin.Default()
	r.GET("/search", searchEndpoint)
	if err = r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func searchEndpoint(c *gin.Context) {
	// Parse request
	query := c.Query("query")
	if query == "" {
		errorResponse(c, http.StatusBadRequest, "Query not specified")
		return
	}
	skip := 0
	take := 10
	if i, err := strconv.Atoi(c.Query("skip")); err == nil {
		skip = i
	}
	if i, err := strconv.Atoi(c.Query("take")); err == nil {
		take = i
	}
	// Perform search
	esQuery := elastic.NewMultiMatchQuery(query, "MPO_REFERENCE", "SERIAL_NUMBER").
		Fuzziness("0").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(skip).Size(take).
		Do(c.Request.Context())
	if err != nil {
		log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		return
	}
	res := SearchResponse{
		Time: fmt.Sprintf("%d", result.TookInMillis),
		Hits: fmt.Sprintf("%d", result.Hits.TotalHits),
	}
	// Transform search results before returning them
	docs := make([]DocumentResponse, 0)
	for _, hit := range result.Hits.Hits {
		var doc DocumentResponse
		json.Unmarshal(*hit.Source, &doc)
		docs = append(docs, doc)
	}
	res.Documents = docs
	c.JSON(http.StatusOK, res)
}

func errorResponse(c *gin.Context, code int, err string) {
	c.JSON(code, gin.H{
		"error": err,
	})
}
