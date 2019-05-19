package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/api/cloudbuild/v1"
)

// BuildStatus referes to Cloud Build status.
type BuildStatus int

const (
	UnknownBuildStatus BuildStatus = iota
	QueuedBuildStatus
	WorkingBuildStatus
	SuccessBuildStatus
	FailureBuildStatus
	InternalErrorBuildStatus
	TimeoutBuildStatus
	CanceledBuildStatus
)

type BuildResult struct {
	Id     string
	Status BuildStatus
	LogURL string
}

func GetBuild(project, id string) (BuildResult, error) {
	res := BuildResult{
		Status: UnknownBuildStatus,
		Id:     id,
	}
	ctx := context.Background()
	svc, err := cloudbuild.NewService(ctx)
	if err != nil {
		return res, err
	}
	bldSvc := cloudbuild.NewProjectsBuildsService(svc)
	b, err := bldSvc.Get(project, id).Do()
	if err != nil {
		return res, err
	}
	res.Status = toBuildStatus(b.Status)
	res.LogURL = b.LogUrl
	return res, nil
}

// CreateBuild creates and starts a CloudBuild on GCP.
func CreateBuild(project, name, bucket, object string, tags []string) (BuildResult, error) {
	res := BuildResult{
		Status: UnknownBuildStatus,
	}
	ctx := context.Background()
	svc, err := cloudbuild.NewService(ctx)
	if err != nil {
		return res, err
	}
	bldSvc := cloudbuild.NewProjectsBuildsService(svc)
	images := generateImageNames(project, name, tags)
	b := cloudbuild.Build{
		Images: images,
		Source: &cloudbuild.Source{
			StorageSource: &cloudbuild.StorageSource{
				Bucket: bucket,
				Object: object,
			},
		},
		Timeout: "1200s",
		Steps: []*cloudbuild.BuildStep{
			&cloudbuild.BuildStep{
				Name: "gcr.io/cloud-builders/docker",
				Args: generateArgs(images),
			},
		},
	}
	r, err := bldSvc.Create(project, &b).Do()
	if err != nil {
		return res, err
	}
	if r.Error != nil {
		return res, errors.New(r.Error.Message)
	}
	var meta cloudbuild.BuildOperationMetadata
	err = json.Unmarshal(r.Metadata, &meta)
	if err != nil {
		return res, err
	}
	res.Status = toBuildStatus(meta.Build.Status)
	res.Id = meta.Build.Id
	res.LogURL = meta.Build.LogUrl
	return res, nil
}

func toBuildStatus(status string) BuildStatus {
	switch status {
	case "STATUS_UNKNOWN":
		return UnknownBuildStatus
	case "QUEUED":
		return QueuedBuildStatus
	case "WORKING":
		return WorkingBuildStatus
	case "SUCCESS":
		return SuccessBuildStatus
	case "FAILURE":
		return FailureBuildStatus
	case "INTERNAL_ERROR":
		return InternalErrorBuildStatus
	case "TIMEOUT":
		return TimeoutBuildStatus
	case "CANCELLED":
		return CanceledBuildStatus
	default:
		return UnknownBuildStatus
	}
}

func generateImageNames(project, name string, tags []string) []string {
	var images []string
	for _, tag := range tags {
		images = append(images, fmt.Sprintf("gcr.io/%v/%v:%v", project, name, tag))
	}
	return images
}

func generateArgs(images []string) []string {
	var args []string
	args = append(args, "build")
	args = append(args, "-f")
	args = append(args, "Dockerfile")
	for _, img := range images {
		args = append(args, "-t")
		args = append(args, img)
	}
	return append(args, ".")
}
