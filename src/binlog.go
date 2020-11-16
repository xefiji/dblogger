package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/iancoleman/strcase"
	"github.com/joho/godotenv"
	"github.com/siddontang/go-mysql/canal"
)

const (
	READABLE_DEFAULT = "EventOccured"
	READABLE_CREATED = "WasCreated"
	READABLE_UPDATED = "HasChanged"
	READABLE_REMOVED = "WasRemoved"
)

type binlogHandler struct {
	canal.DummyEventHandler
	messenger *Rmq
}

type event struct {
	Readable string
	Action   string
	Schema   string
	Table    string
	Payload  map[string]interface{}
	Origin   map[string]interface{}
	Header   eventHeader
}

type eventHeader struct {
	Timestamp uint32
	EventType string
	ServerID  uint32
	EventSize uint32
}

//update User set name = concat("John", char(round(rand()*25)+97)) where id = 3;
//insert into User(name, status) values(concat("FX", char(round(rand()*25)+97)), "active");

//run runs the binlog listener app
func run(exchange string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//connect to db
	c, err := getDefaultCanal()
	if err == nil {
		coords, err := c.GetMasterPos()
		if err == nil {

			//connect to broker
			client := Rmq{}
			client.Connect(exchange)

			b := &binlogHandler{
				messenger: &client,
			}

			c.SetEventHandler(b)
			c.RunFrom(coords)
		}
	}
}

//OnRow
func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {

	defer func() {
		if r := recover(); r != nil {
			fmt.Print(r, " ", string(debug.Stack()))
		}
	}()

	//hydrate events struct
	dbEvent := event{
		getReadable(e),
		e.Action,
		e.Table.Schema,
		e.Table.Name,
		getPayload(e),
		getOrigin(e),
		eventHeader{
			e.Header.Timestamp,
			e.Header.EventType.String(),
			e.Header.ServerID,
			e.Header.EventSize,
		},
	}

	encoded, err := json.Marshal(dbEvent)
	if err != nil {
		return err
	}

	h.messenger.Send(h.messenger.exchange, encoded)

	return nil
}

//getDefaultCanal sets a new canal with database config
func getDefaultCanal() (*canal.Canal, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	cfg.User = os.Getenv("DB_USER")
	cfg.Password = os.Getenv("DB_PASSWORD")
	cfg.Flavor = os.Getenv("DB_FLAVOR")

	cfg.Dump.ExecutionPath = ""

	return canal.NewCanal(cfg)
}

//getReadable returns a human readable string representation of the event, based on table name
func getReadable(e *canal.RowsEvent) string {
	readable := fmt.Sprintf("%s%s", strcase.ToCamel(e.Table.Name), READABLE_DEFAULT)
	switch e.Action {
	case canal.InsertAction:
		readable = fmt.Sprintf("%s%s", strcase.ToCamel(e.Table.Name), READABLE_CREATED)
	case canal.UpdateAction:
		readable = fmt.Sprintf("%s%s", strcase.ToCamel(e.Table.Name), READABLE_UPDATED)
	case canal.DeleteAction:
		readable = fmt.Sprintf("%s%s", strcase.ToCamel(e.Table.Name), READABLE_REMOVED)
	}
	return readable
}

//getOrigin returns the datas used in the query
func getPayload(e *canal.RowsEvent) map[string]interface{} {
	payload := make(map[string]interface{})

	//default payload index is the first one, except on update where
	//first one is original datas, second one is new datas
	payloadIndex := 0
	if e.Action == canal.UpdateAction && len(e.Rows) == 2 {
		payloadIndex = 1
	}

	//hydrate payload
	for i, value := range e.Rows[payloadIndex] {
		payload[e.Table.Columns[i].Name] = value
	}

	return payload
}

//getOrigin returns the original datas in case of update (else empty map)
func getOrigin(e *canal.RowsEvent) map[string]interface{} {
	origin := make(map[string]interface{})

	//default payload index is the first one, except on update where
	//first one is original datas, second one is new datas
	if e.Action == canal.UpdateAction && len(e.Rows) == 2 {
		//hydrate origin
		for i, value := range e.Rows[0] {
			origin[e.Table.Columns[i].Name] = value
		}
	}

	return origin
}

//pretty marshalls sruct and prints prettyfied version for debug purpose
func pretty(element interface{}) {
	json, _ := json.MarshalIndent(element, "", "  ")
	fmt.Printf("\n\n%+v\n\n", string(json))
}
