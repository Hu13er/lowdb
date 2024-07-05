package lowdb

import (
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

type KVStoreHTTP struct {
	Store KVStore
}

func (kvHttp *KVStoreHTTP) Handlers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", kvHttp.getKeysHandler)
	mux.HandleFunc("GET /{key}/", kvHttp.getHandler)
	mux.HandleFunc("POST /{key}/", kvHttp.setHandler)
	mux.HandleFunc("DELETE /{key}/", kvHttp.setHandler)
	return mux
}

func (kvHttp *KVStoreHTTP) getKeysHandler(w http.ResponseWriter, _ *http.Request) {
	keys, err := kvHttp.Store.Keys()
	if err != nil {
		log.Panicln(err)
	}
	io.WriteString(w,
		strings.Join(keys, "\n"))
}

func (kvHttp *KVStoreHTTP) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if len(key) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data, err := kvHttp.Store.Get(key)
	if err != nil {
		log.Panicln(err)
	}
	if data.Empty() {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	r.Header.Add("LOWDB-key", data.Key)
	r.Header.Add("LOWDB-revision", strconv.Itoa(data.Revision))
	r.Header.Add("LOWDB-created_at", data.CreatedAt.Format(time.ANSIC))
	for k, vs := range data.Headers {
		for _, v := range vs {
			r.Header.Add("LOWDB-META-"+k, v)
		}
	}
	w.Write(data.Value)
}

func (kvHttp *KVStoreHTTP) setHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if len(key) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rev := -1
	if revStr := r.Header.Get("LOWDB-revision"); len(revStr) > 0 {
		var err error
		rev, err = strconv.Atoi(revStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	headers := make(map[string][]string)
	for k, vs := range r.Header {
		trimedk := strings.TrimPrefix(k, "LOWDB-META-")
		if trimedk == k {
			continue
		}
		headers[trimedk] = slices.Clone(vs)
	}

	value, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch err := kvHttp.Store.Set(KeyValueMetadata{
		Key:      key,
		Value:    value,
		Revision: rev,
		Headers:  headers,
	}); err {
	case nil:
		w.WriteHeader(http.StatusOK)
		return
	case InvalidRevsion:
		w.WriteHeader(http.StatusConflict)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (kvHttp *KVStoreHTTP) deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if len(key) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rev := -1
	if revStr := r.Header.Get("LOWDB-revision"); len(revStr) > 0 {
		var err error
		rev, err = strconv.Atoi(revStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	switch err := kvHttp.Store.Delete(KeyValueMetadata{
		Key:      key,
		Revision: rev,
	}); err {
	case nil:
		w.WriteHeader(http.StatusOK)
		return
	case InvalidRevsion:
		w.WriteHeader(http.StatusConflict)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
