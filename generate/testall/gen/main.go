package main

import (
	"github.com/marshome/apis/generate/testdata/acceleratedmobilepageurl/v1/gen/server/api"
	"github.com/marshome/apis/restful"
)

func main() {
	acceleratedmobilepageurl_v1.RouteAmpUrlsService(restful.NewRouter(), &acceleratedmobilepageurl_v1.DefaultAmpUrlsService{})
}
