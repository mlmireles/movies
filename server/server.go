package server

import (
	"io/ioutil"
	"log"
	"movies/keys"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// PathPrefix for server api
const PathPrefix = "/v1/movies"

// RegisterHandlers create all server api handlers
func RegisterHandlers() http.Handler {
	return NewRouter()
}

// NewRouter returns new mux router with given handlers
func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(PathPrefix, errorHandler(ListMoviesDiscover)).Methods(http.MethodGet)
	router.HandleFunc(PathPrefix+"/{movie_id}", errorHandler(GetMovie)).Methods(http.MethodGet)

	return router
}

// badRequest is handled by setting the status code in the reply to StatusBadRequest.
type badRequest struct{ error }

// notFound is handled by setting the status code in the reply to StatusNotFound.
type notFound struct{ error }

// notAuthorized is handled by setting the status code in the reply to StatusUnauthorized
type notAuthorized struct{ error }

// errorHandler wraps a function returning an error by handling the error and
// returning a http.Handler.
// If the error is of the one of the types defined above, it is handled as
// described for every type.
// If the error is of another type, it is considered as an internal error and
// its message is logged.
func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		switch err.(type) {
		case badRequest:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case notFound:
			http.Error(w, "Task not found", http.StatusNotFound)
		case notAuthorized:
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
		default:
			log.Println(err)
			http.Error(w, "oops", http.StatusInternalServerError)
		}
	}
}

// newRequest generic request function
// m httpMethod
func newRequest(m string, url string, w http.ResponseWriter) error {
	req, err := http.NewRequest(m, url, nil)
	if err != nil {
		return badRequest{}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return notFound{}
	} else if res.StatusCode == http.StatusUnauthorized {
		return notAuthorized{}
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	w.WriteHeader(res.StatusCode)
	w.Write(body)
	return nil
}

// ListMoviesDiscover get a list for dicover movies
func ListMoviesDiscover(w http.ResponseWriter, r *http.Request) error {
	url := keys.PATH_API_TMD + "discover/movie?" + "api_key=" + keys.API_KEY
	params := r.URL.Query()
	for n, v := range params {
		url += "&" + n + "=" + v[0]
	}

	return newRequest(http.MethodGet, url, w)
}

// GetMovie request single movie info given the id
func GetMovie(w http.ResponseWriter, r *http.Request) error {
	var id string
	s := strings.Split(r.URL.Path, "/")
	for i := range s {
		id = s[i]
	}

	url := keys.PATH_API_TMD + "movie/" + id + "?api_key=" + keys.API_KEY

	return newRequest(http.MethodGet, url, w)
}
