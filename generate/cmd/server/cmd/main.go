package main

import (
	"github.com/marshome/apis/generate/cmd/server"
	"github.com/marshome/apis/restful"
)

func main() {
	api.RouteProjectsService(restful.NewRouter(nil, nil), &api.DefaultProjectsService{})
}