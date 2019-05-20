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
	// UnknownBuildStatus refers to an edge case wherein the status is unknown.
	UnknownBuildStatus BuildStatus = iota
	// QueuedBuildStatus means the build is waiting in the queue to be processed.
	QueuedBuildStatus
	// WorkingBuildStatus means the build is in progress.
	WorkingBuildStatus
	// SuccessBuildStatus means the build has completed successfully.
	SuccessBuildStatus
	// FailureBuildStatus means the build ended because of user failure.
	FailureBuildStatus
	// InternalErrorBuildStatus means the build ended because of an error in the build system.
	InternalErrorBuildStatus
	// TimeoutBuildStatus means the build ended because it ran too long.
	TimeoutBuildStatus
	// CanceledBuildStatus means the build ended because a user has canceled it.
	CanceledBuildStatus
)

// BuildResult represents build info returned by build operations.
type BuildResult struct {
	ID     string
	Status BuildStatus
	LogURL string
}

// GetBuild retrieves the latest build status for a given GCP Cloud Build.
func GetBuild(project, id string) (BuildResult, error) {
	res := BuildResult{
		Status: UnknownBuildStatus,
		ID:     id,
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
	res.ID = meta.Build.Id
	res.LogURL = meta.Build.LogUrl
	return res, nil
}

// Convert build status strings into a BuildStatus enum.
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

// Generates full GCP docker image names.
func generateImageNames(project, name string, tags []string) []string {
	var images []string
	for _, tag := range tags {
		images = append(images, fmt.Sprintf("gcr.io/%v/%v:%v", project, name, tag))
	}
	return images
}

// Generates CloudBuild arguments for building docker images.
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
