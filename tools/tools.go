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
	apihandlers.PanicIfNil(err)
	fmt.Fprintf(w, "%s", string(jsonstr))
}
