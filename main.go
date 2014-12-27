/*
The MIT License (MIT)

Copyright (c) 2014 Richard Lyman

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// bob is a Binary OBject store REST wrapper around the GET and SET(NX) commands in Redis.
package main

import (
	"flag"
	"fmt"
	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

var lockVersions = flag.Bool("lockVersions", false, "Switch for locking versions. Content associated with a locked version can not be rewritten.")
var hostPort = flag.String("hostPort", ":8080", "The host and port to bind to.")
var redisHostPort = flag.String("redisHostPort", "127.0.0.1:6379", "The redis host and port to bind to.")

var p *pool.Pool

func main() {
	flag.Parse()
	tmp, err := pool.NewPool("tcp", *redisHostPort, 10)
	if err != nil {
		log.Fatal("Failed to create redis connection pool:", err)
	}
	p = tmp
	r := mux.NewRouter()
	r.HandleFunc("/{v}", get).Methods("GET")
	r.HandleFunc("/{v}", post).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) })
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(*hostPort, nil))
}

func get(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)["v"]
	c, err := p.Get()
	if err != nil {
		log.Println("Unable to get client from redis connection pool:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer p.Put(c)
	response := c.Cmd("GET", v)
	if response.Type == redis.NilReply {
		log.Printf("Failed to get content for version '%s' (NilReply): %s\n", v, err)
		w.WriteHeader(http.StatusNotFound)
	} else if response.Err != nil {
		log.Printf("Failed to get content for version '%s': %s\n", v, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		if b, err := response.Bytes(); err != nil {
			log.Printf("Unable to convert redis response to bytes for GET on version '%s': %s\n", v, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			fmt.Fprint(w, b)
		}
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)["v"]
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("Failed to read POSTed request body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := p.Get()
	if err != nil {
		log.Println("Unable to get client from redis connection pool:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer p.Put(c)
	if *lockVersions {
		response := c.Cmd("SET", v, b, "NX")
		if response.Type == redis.NilReply {
			log.Printf("Failed to set content for version '%s' (NilReply): %s\n", v, err)
			w.WriteHeader(http.StatusBadRequest)
		} else if response.Err != nil {
			log.Printf("Failed to set content for version '%s': %s\n", v, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		if err = c.Cmd("SET", v, b).Err; err != nil {
			log.Println("Failed to set content:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
