package eio

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

func CreateFileJson(items interface{}, jsonFile string) (file *os.File, err error) {
	var plantsJSON []byte
	if plantsJSON, err = json.MarshalIndent(items, "", "    "); err != nil {
		log.Fatal("Cannot marshal JSON", err)
		return
	}

	if file, err = os.Create(jsonFile); err != nil {
		log.Fatal("Cannot create file", err)
		return
	}
	log.Println("file written", jsonFile)
	defer file.Close()
	file.Write(plantsJSON)

	return
}

/*
	//t := reflect.TypeOf(r.factoryFn())
	//items := reflect.ArrayOf(0, t)
	//v := reflect.New(items)
	//p := v.Interface()


	sliceType := reflect.SliceOf(reflect.TypeOf(r.factoryFn()))
	slicePtr := reflect.New(sliceType)

	err = json.Unmarshal(bytes, slicePtr.Interface())

 */

func LoadArrayJsonByReflect(jsonFile string, t reflect.Type) (ret []interface{}, err error) {
	var filename *os.File
	if filename, err = os.Open(jsonFile); err != nil {
		return
	}
	defer filename.Close()

	var data []byte
	if data, err = ioutil.ReadAll(filename); err != nil {
		return
	}

	var tmp []json.RawMessage
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}

	i := 0
	items := make([]interface{}, len(tmp))
	for _, raw := range tmp {
		v := reflect.New(t.Elem())
		newP := v.Interface()

		if err = json.Unmarshal(raw, newP); err != nil {
			return
		}
		items[i] = newP
		i += 1
	}
	ret = items
	return
}
