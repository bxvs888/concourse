package exec

import (
	"archive/tar"
	"io"
	"os"

	"github.com/concourse/atc/resource"
)

type resourceStep struct {
	Session resource.Session

	Delegate ResourceDelegate

	Tracker resource.Tracker
	Type    resource.ResourceType

	Action func(resource.Resource, ArtifactSource) resource.VersionedSource

	PreviousStep Step
	Repository   *SourceRepository

	Resource        resource.Resource
	VersionedSource resource.VersionedSource
}

func (step resourceStep) Using(prev Step, repo *SourceRepository) Step {
	step.PreviousStep = prev
	step.Repository = repo

	return failureReporter{
		Step:          &step,
		ReportFailure: step.Delegate.Failed,
	}
}

func (ras *resourceStep) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	resource, err := ras.Tracker.Init(ras.Session, ras.Type)
	if err != nil {
		return err
	}

	ras.Resource = resource
	ras.VersionedSource = ras.Action(resource, ras.Repository)

	err = ras.VersionedSource.Run(signals, ready)
	if err != nil {
		return err
	}

	ras.Delegate.Completed(VersionInfo{
		Version:  ras.VersionedSource.Version(),
		Metadata: ras.VersionedSource.Metadata(),
	})

	return nil
}

func (ras *resourceStep) Release() error {
	if ras.Resource != nil {
		return ras.Resource.Destroy()
	}

	return nil
}

func (ras *resourceStep) Result(x interface{}) bool {
	switch v := x.(type) {
	case *VersionInfo:
		*v = VersionInfo{
			Version:  ras.VersionedSource.Version(),
			Metadata: ras.VersionedSource.Metadata(),
		}
		return true

	default:
		return false
	}
}

type fileReadCloser struct {
	io.Reader
	io.Closer
}

func (ras *resourceStep) StreamTo(destination ArtifactDestination) error {
	out, err := ras.VersionedSource.StreamOut(".")
	if err != nil {
		return err
	}

	return destination.StreamIn(".", out)
}

func (ras *resourceStep) StreamFile(path string) (io.ReadCloser, error) {
	out, err := ras.VersionedSource.StreamOut(path)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(out)

	_, err = tarReader.Next()
	if err != nil {
		return nil, ErrFileNotFound
	}

	return fileReadCloser{
		Reader: tarReader,
		Closer: out,
	}, nil
}
