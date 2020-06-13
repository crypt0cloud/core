package model_gae

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"

	"github.com/crypt0cloud/core/tools"
)

func GetClient(r *http.Request) *http.Client {
	ctx := tools.Context(r)
	ctx, _ = context.WithTimeout(ctx, 60*time.Second)
	return urlfetch.Client(ctx)
}
