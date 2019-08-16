package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/onlyangel/apihandlers"
	"google.golang.org/appengine"
	"html"
	"net/http"
)

func Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func FormValueEscaped(r *http.Request, variable string) string {
	str := r.FormValue(variable)
	return html.EscapeString(str)
}

func PrintJson(w http.ResponseWriter, v interface{}) {
	jsonstr, err := json.Marshal(v)
	apihandlers.PanicIfNotNil(err)
	fmt.Fprintf(w, "%s", string(jsonstr))
}

func API_Error(payload []byte) (bool, string) {
	obj := new(apihandlers.ErrorType)
	err := json.Unmarshal(payload, obj)
	apihandlers.PanicIfNotNil(err)
	return obj.Error != "", obj.Error
}

func IsStringArrayEqual(a, b []string) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func IsStringInArray(original string, arr []string) bool {
	sino := false
	for i := range arr {
		if original == arr[i] {
			return true
		}
	}
	return !sino
}
