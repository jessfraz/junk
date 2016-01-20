package server

import (
	"github.com/jfrazelle/hulk/api/grpc/types"
	"golang.org/x/net/context"
)

type apiServer struct {
	ArtifactsDir string
	StateDir     string
}

// NewServer returns grpc server instance
func NewServer(artifactsDir, stateDir string) types.APIServer {
	return &apiServer{
		ArtifactsDir: artifactsDir,
		StateDir:     stateDir,
	}
}

func (s *apiServer) StartJob(ctx context.Context, c *types.StartJobRequest) (*types.StartJobResponse, error) {
	return &types.StartJobResponse{
		Id: uint32(5),
	}, nil
}

func (s *apiServer) DeleteJob(ctx context.Context, r *types.DeleteJobRequest) (*types.DeleteJobResponse, error) {
	return &types.DeleteJobResponse{}, nil
}

func (s *apiServer) ListJobs(ctx context.Context, r *types.ListJobsRequest) (*types.ListJobsResponse, error) {
	return &types.ListJobsResponse{}, nil
}

func (s *apiServer) State(ctx context.Context, r *types.StateRequest) (*types.StateResponse, error) {
	return &types.StateResponse{}, nil
}

func (s *apiServer) Logs(r *types.LogsRequest, stream types.API_LogsServer) error {
	return nil
}
