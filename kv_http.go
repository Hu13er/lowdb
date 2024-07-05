package lowdb

import (
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
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
	io.WriteString(w,
		strings.Join(kvHttp.Store.Keys(), "\n"))
}

func (kvHttp *KVStoreHTTP) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if len(key) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data := kvHttp.Store.Get(key)
	if data.Empty() {
		w.WriteHeader(http.StatusNotFound)
		return
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
	if revStr := r.Header.Get("rivision"); len(revStr) > 0 {
		var err error
		rev, err = strconv.Atoi(revStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	value, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	headers := make(map[string][]string)
	for k, vs := range r.Header {
		trimedk := strings.TrimPrefix(k, "LOWDB-")
		if trimedk == k {
			continue
		}
		headers[trimedk] = slices.Clone(vs)
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
	if revStr := r.Header.Get("rivision"); len(revStr) > 0 {
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
