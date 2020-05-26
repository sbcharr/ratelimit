package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sbcharr/ratelimit"
	"golang.org/x/xerrors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	errTooManyRequests   = xerrors.New("you are sending too many requests, please slow down")
	errBurstLimitReached = xerrors.New("burst limit exceeded")
)

// Response contains the info on the HTTP response back to the caller
type Response struct {
	Status      int    `json:"status"`
	Description string `json:"description"`
}

// WebAppAPIServer is the API server to receive requests from users
// This is the API server facing the external world
func WebAppAPIServer(rl ratelimit.APIRateLimiter) {
	router := mux.NewRouter()
	//router.HandleFunc("/v1/list_instances", listInstances)
	router.HandleFunc("/v1/check_ratelimit/{key}", func(w http.ResponseWriter, request *http.Request) {
		checkRateLimit(w, request, rl)
	})
	router.HandleFunc("/v1", hello)
	http.Handle("/", router)

	if err := http.ListenAndServe(":80", router); err != nil {
		log.Fatal(fmt.Sprintf(`WebAppAPIServer() ! %s`, err.Error()))
	}
}

func hello(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func checkRateLimit(w http.ResponseWriter, request *http.Request, rl ratelimit.APIRateLimiter) {
	// fmt.Println(fmt.Sprintf("%s /v1/check_ratelimit", request.Method))
	ctx := context.TODO()
	switch request.Method {
	case "POST":
		if err := checkRateLimitHelper(ctx, w, request, rl); err != nil {
			fmt.Println(err.Error())
			if err.Error() == errTooManyRequests.Error() {
				writeJSONResponse(w, http.StatusTooManyRequests, xerrors.New("sending too many requests, please slow down").Error())
			} else if err.Error() == errBurstLimitReached.Error() {
				writeJSONResponse(w, http.StatusTooManyRequests, xerrors.New("request burst limit exceeded, please slow down").Error())
			} else {
				writeJSONResponse(w, http.StatusInternalServerError, errors.New("internal server error, try after sometime").Error())
			}
		}
		// fmt.Println("createJobHelpe() returned")
	default:
		writeJSONResponse(w, http.StatusMethodNotAllowed, "Allowed Methods: POST")
	}

}

func checkRateLimitHelper(ctx context.Context, w http.ResponseWriter, request *http.Request, rl ratelimit.APIRateLimiter) error {
	vars := mux.Vars(request)
	key := vars["key"]
	err := rl.RunContext(ctx, key)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	return nil
}

func writeJSONResponse(w http.ResponseWriter, status int, description string) {
	resp := Response{Status: status, Description: description}
	msg, err := json.Marshal(resp)
	if err != nil {
		log.Fatal("debug1:", err.Error())
	}
	// fmt.Println("Returning response: " + string(msg))
	w.WriteHeader(status)
	http.Error(w, string(msg), status)
}
