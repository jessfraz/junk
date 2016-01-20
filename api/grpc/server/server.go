package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/jfrazelle/hulk/api/grpc/types"
	"golang.org/x/net/context"
)

var (
	dbfile       = "bolt.db"
	jobsDBBucket = []byte("jobs")
	stateSuffix  = "_state"
)

type apiServer struct {
	ArtifactsDir string
	StateDir     string
	DB           *bolt.DB
}

// State defines the statuses of a job.
type State []byte

var (
	jobCreated   State = []byte("created")
	jobCompleted State = []byte("completed")
	jobFailed    State = []byte("failed")
	jobRunning   State = []byte("running")
	jobStarted   State = []byte("started")
)

// NewServer returns grpc server instance
func NewServer(artifactsDir, stateDir string) (types.APIServer, error) {
	if err := os.MkdirAll(stateDir, 0666); err != nil {
		return nil, fmt.Errorf("attempt to create state directory %s failed: %v", stateDir, err)
	}

	// delete any leftover db and start fresh
	dbpath := filepath.Join(stateDir, dbfile)
	if err := os.RemoveAll(dbpath); err != nil {
		return nil, fmt.Errorf("attempt to remove %s failed: %v", dbpath, err)
	}

	db, err := bolt.Open(dbpath, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("Opening database at %s failed: %v", dbpath, err)
	}

	// create the jobs bucket
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucket(jobsDBBucket); err != nil {
			return fmt.Errorf("Creating bucket %s failed: %v", jobsDBBucket, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &apiServer{
		ArtifactsDir: artifactsDir,
		StateDir:     stateDir,
		DB:           db,
	}, nil
}

func jobIDByte(id uint32) []byte {
	return []byte(strconv.Itoa(int(id)))
}

func jobStateByte(id uint32) []byte {
	return []byte(strconv.Itoa(int(id)) + stateSuffix)
}

func (s *apiServer) updateState(id uint32, state State) error {
	if err := s.DB.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(jobsDBBucket).Put(jobStateByte(id), state)
	}); err != nil {
		return fmt.Errorf("Updating state for %d to %s failed: %v", id, state, err)
	}

	return nil
}

func (s *apiServer) StartJob(ctx context.Context, c *types.StartJobRequest) (*types.StartJobResponse, error) {
	job := types.Job{
		Name:      c.Name,
		Args:      c.Args,
		Artifacts: c.Artifacts,
	}

	addJob := func(tx *bolt.Tx) error {
		// Retrieve the jobs bucket.
		b := tx.Bucket(jobsDBBucket)

		// Generate ID for the job.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so ignore the error check.
		id, _ := b.NextSequence()
		job.Id = uint32(id)

		// Marshal job data into bytes.
		buf, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("Marshal job failed: %v", err)
		}

		// put the job into the bucket
		if err := b.Put(jobIDByte(job.Id), buf); err != nil {
			return fmt.Errorf("Putting job into bucket failed: %v", err)
		}
		// put the job state into the bucket
		// we will keep these seperate for updating more easily
		if err := b.Put(jobStateByte(job.Id), jobCreated); err != nil {
			return fmt.Errorf("Putting job state into bucket failed: %v", err)
		}
		return nil
	}

	if err := s.DB.Update(addJob); err != nil {
		return nil, fmt.Errorf("Adding job to database failed: %v", err)
	}

	// create the job runner
	j, err := createJob(job.Id, s.StateDir, job.Args)
	if err != nil {
		return nil, err
	}

	// start the command
	if err := j.cmd.Start(); err != nil {
		return nil, fmt.Errorf("Starting cmd [%s] failed: %v", j.cmdStr, err)
	}
	if err := s.updateState(job.Id, jobStarted); err != nil {
		return nil, err
	}

	// run and wait for the command, in a go routine
	go func() {
		state := jobCompleted
		if err := j.run(); err != nil {
			logrus.Error(err)
			state = jobFailed
		}
		if err := s.updateState(job.Id, state); err != nil {
			logrus.Error(err)
		}
	}()

	if err := s.updateState(job.Id, jobRunning); err != nil {
		return nil, err
	}

	return &types.StartJobResponse{
		Id: job.Id,
	}, nil
}

func (s *apiServer) DeleteJob(ctx context.Context, r *types.DeleteJobRequest) (*types.DeleteJobResponse, error) {
	// delete the artifacts, if any
	artifacts := filepath.Join(s.ArtifactsDir, string(jobIDByte(r.Id)))
	if err := os.RemoveAll(artifacts); err != nil {
		return nil, fmt.Errorf("attempt to remove %s failed: %v", artifacts, err)
	}

	if err := s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(jobsDBBucket)
		// delete the job
		if err := b.Delete(jobIDByte(r.Id)); err != nil {
			return fmt.Errorf("Deleting job from bucket failed: %v", err)
		}

		// delete the job state
		if err := b.Delete(jobStateByte(r.Id)); err != nil {
			return fmt.Errorf("Deleting job state from bucket failed: %v", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Deleting job id %d from db failed: %v", r.Id, err)
	}

	return &types.DeleteJobResponse{}, nil
}

func (s *apiServer) ListJobs(ctx context.Context, r *types.ListJobsRequest) (*types.ListJobsResponse, error) {
	var jobs []*types.Job
	if err := s.DB.View(func(tx *bolt.Tx) error {
		// Retrieve the jobs bucket.
		b := tx.Bucket(jobsDBBucket)

		return b.ForEach(func(k, v []byte) error {
			// ignore state keys
			if strings.HasSuffix(string(k), stateSuffix) {
				return nil
			}

			var job types.Job
			if err := json.Unmarshal(v, &job); err != nil {
				return fmt.Errorf("Unmarshal job failed: %v\nRaw: %s", err, v)
			}

			job.Status = string(b.Get(jobStateByte(job.Id)))

			jobs = append(jobs, &job)

			return nil
		})
	}); err != nil {
		return nil, fmt.Errorf("Getting jobs from db failed: %v", err)
	}
	return &types.ListJobsResponse{Jobs: jobs}, nil
}

func (s *apiServer) State(ctx context.Context, r *types.StateRequest) (*types.StateResponse, error) {
	var state []byte
	if err := s.DB.View(func(tx *bolt.Tx) error {
		state = tx.Bucket(jobsDBBucket).Get(jobStateByte(r.Id))

		return nil
	}); err != nil {
		return nil, fmt.Errorf("Getting jobs from db failed: %v", err)
	}
	return &types.StateResponse{
		Status: string(state),
	}, nil
}

func (s *apiServer) Logs(r *types.LogsRequest, stream types.API_LogsServer) error {

	return nil
}
