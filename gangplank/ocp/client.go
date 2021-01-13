package ocp

import (
	"os"

	log "github.com/sirupsen/logrus"

	workspec "github.com/coreos/gangplank/builder-workspec"
	"github.com/coreos/gangplank/clustercontext"
	"github.com/coreos/gangplank/constants"
	"github.com/coreos/gangplank/errors"
)

// Builder implements the Build
type Builder interface {
	Exec(ctx clustercontext.ClusterContext) error
}

var (
	// cosaSrvDir is where the build directory should be. When the build API
	// defines a contextDir then it will be used. In most cases this should be /srv
	cosaSrvDir = constants.DefaultContextDir
)

// NewBuilder returns a Builder. NewBuilder determines what
// "Builder" to return.
func NewBuilder(ctx clustercontext.ClusterContext) (Builder, error) {
	inCluster := true
	if _, ok := os.LookupEnv(localPodEnvVar); ok {
		log.Infof("EnvVar %s defined, using local pod mode", localPodEnvVar)
		inCluster = false
	}

	ws, err := workspec.NewBuilder(ctx)
	if err == nil {
		return ws, nil
	}
	bc, err := newBC(ctx, &clustercontext.Cluster{inCluster: inCluster})
	if err == nil {
		return bc, nil
	}
	return nil, errors.ErrNoWorkFound
}
