package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
)

const DEFAULT_INDEX = "binlog_events"

var (
	instance *es
	once     sync.Once
	raw      map[string]interface{}
)

type Es interface {
	Get() *elasticsearch.Client
	Map(indexName string, fileName string)
	IndexOne(indexName string, id string, indexDatas io.Reader) error
	IndexBulk(indexName string, buffer bytes.Buffer) error
	IndexExists(indexName string) bool
}

type es struct {
	Client *elasticsearch.Client
}

//GetElasticInstance instantiate a new elastic client, makes some checks on connection availability
//then hydrates a new instance of es struct with client and returns it
//or returns the already instantiated singleton
func GetESInstance() *es {
	once.Do(func() {
		client, err := elasticsearch.NewDefaultClient()
		if err != nil {
			log.Panicf("❌ [elastic] Error creating the client: %s", err)
		}

		res, err := client.Info()
		if err != nil {
			log.Panicf("❌ [elastic] Error getting response: %s", err)
		}

		if res.IsError() || res.StatusCode != 200 {
			log.Panicf("❌ [elastic] Status code not satisfying: %d", res.StatusCode)
		}

		log.Printf("[elastic] %s ✅", res.Status())

		instance = &es{
			Client: client,
		}
	})

	return instance
}

//Get returns elastic client from singleton
func (es *es) Get() *elasticsearch.Client {
	return es.Client
}

//Map pushes mapping to elastic server, based on file path
func (es *es) Map(indexName string, fileName string) {
	//Delete index first
	if _, err := es.Client.Indices.Delete([]string{indexName}); err != nil {
		log.Fatalf("❌ [elastic] Cannot delete index: %s", err)
	}

	//creates
	res, err := es.Client.Indices.Create(indexName)
	if err != nil {
		log.Fatalf("❌ [elastic] Cannot create index: %s", err)
	}
	if res.IsError() {
		log.Fatalf("❌ [elastic] Cannot create index: %s", res)
	}

	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("❌ [elastic] Cannot open mapping file: %s", err)
	}

	res, err = es.Client.Indices.PutMapping(
		[]string{indexName},
		strings.NewReader(string(f)),
	)
	if err != nil {
		log.Fatalf("❌ [elastic] Cannot put mapping: %s", err)
	}

	defer res.Body.Close()

	res, err = es.Client.Indices.GetMapping(es.Client.Indices.GetMapping.WithIndex(indexName))
	if err != nil {
		log.Fatalf("❌ [elastic] Error getting the response: %s", err)
	}

	log.Printf("[elastic] indexation ok for %s from %s ✅", indexName, fileName)
}

//IndexExists
func (es *es) IndexExists(indexName string) bool {
	res, err := es.Client.Indices.GetMapping(es.Client.Indices.GetMapping.WithIndex(indexName))
	if err != nil {
		log.Fatalf("❌ [elastic] Error getting the response: %s", err)
	}
	return !res.IsError()
}

//IndexOne runs index on one id
func (es *es) IndexOne(indexName string, id string, indexDatas io.Reader) error {
	res, err := es.Client.Create(indexName, id, indexDatas)
	if nil != res {
		defer res.Body.Close()
		if res.IsError() {
			if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
				log.Fatalf("❌ [elastic] Failure to parse response body: %s", err)
			} else {
				log.Printf("❌ [elastic] %d %s: %s",
					res.StatusCode,
					raw["error"].(map[string]interface{})["type"],
					raw["error"].(map[string]interface{})["reason"],
				)
			}
		} else {
			log.Printf("[elastic] %s ✅\n", res.Status())
		}
	}

	return err
}

//IndexBulk runs index on multiple datas
func (es *es) IndexBulk(indexName string, buffer bytes.Buffer) error {
	return errors.New("stub")
}
