package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

var dbname string = "urlspath.db"

//datavaseProcesses creates a db, bucket, and inserts the k:v pairs from the provided map
func databaseProcesses(dbname string, entries map[string]string) error {
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range entries {
			err = b.Put([]byte(k), []byte(v))
			if err != nil {
				return fmt.Errorf("put key: %s", err)
			}
		}
		return nil

	})

	return nil
}

//dbHandler accesses the k:v pairs from the db and build a map[string]string
//if a key in the map matches r.URL.Path the request is redirected to the value
//e.g the desired URL
//Use of a map in the function assumes that the db k:v are generated outside of this tool
//so there would be no knowledge of that k:v pairs exist.
//Use of a map also allows for using the mapHandler function in
//https://github.com/jeremiahbailey/urlredirect
func dbHandler(URLPaths map[string]string, fallback http.Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var dbv string
		dbvalues := make(map[string]string)
		db, err := bolt.Open(dbname, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("MyBucket"))
			for k := range URLPaths {
				key := k
				dbv = string(b.Get([]byte(key)))
				dbvalues[key] = dbv

			}

			return nil
		})
		for k, v := range dbvalues {
			if r.URL.Path == k {
				http.Redirect(w, r, string(v), http.StatusFound)
			}
		}
	}
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func main() {
	port := ":8000"
	mux := defaultMux()

	URLPaths := make(map[string]string)
	URLPaths["/dbpathname"] = "https://google.com"
	URLPaths["/otherdbpath"] = "https://google.com/robots.txt"

	databaseProcesses(dbname, URLPaths)
	fmt.Printf("Starting the server on port: %v\n", port)
	http.ListenAndServe(port, dbHandler(URLPaths, mux))
}
