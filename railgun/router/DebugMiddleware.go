package router

import (
	"fmt"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"net/http"
)

var DebugMiddleWare = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestObject := map[string]string{
			"url":    r.URL.String(),
			"Method": r.Method,
		}

		requestProperties, err := gss.SerializeString(requestObject, "properties", []string{}, gss.NoLimit)
		if err != nil {
			panic(err)
		}

		fmt.Println("# Request #")
		fmt.Println(requestProperties)
		next.ServeHTTP(w, r)
	})
}
