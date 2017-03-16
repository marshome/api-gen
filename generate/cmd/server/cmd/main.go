package main

import (
	"github.com/marshome/apis/generate/cmd/server"
	"github.com/marshome/apis/restful"
)

type ProjectsService struct {
}

func (s *ProjectsService) AllocateIds(ctx *restful.Context, req *api.AllocateIdsRequest) (resp *api.AllocateIdsResponse, err error) {
	return nil, nil
}

func (s *ProjectsService) BeginTransaction(ctx *restful.Context, req *api.BeginTransactionRequest) (resp *api.BeginTransactionResponse, err error) {
	return nil, nil
}

func (s *ProjectsService) Commit(ctx *restful.Context, req *api.CommitRequest) (resp *api.CommitResponse, err error) {
	return nil, nil
}

func (s *ProjectsService) Lookup(ctx *restful.Context, req *api.LookupRequest) (resp *api.LookupResponse, err error) {
	return nil, nil
}

func (s *ProjectsService)Rollback(ctx *restful.Context, req *api.RollbackRequest) (resp *api.RollbackResponse, err error) {
	return nil, nil
}

func (s *ProjectsService) RunQuery(ctx *restful.Context, req *api.RunQueryRequest) (resp *api.RunQueryResponse, err error) {
	return nil, nil
}


func main() {
	api.RegistProjectsService(restful.NewRouter(nil, nil), &ProjectsService{})
}