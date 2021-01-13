package buildv1

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/coreos/gangplank/pkg/buildv1/bc"
	"github.com/coreos/gangplank/pkg/buildv1/workspec"
	"github.com/coreos/gangplank/pkg/clustercontext"
	"github.com/coreos/gangplank/pkg/errors"
	"github.com/coreos/gangplank/pkg/shared"
	"github.com/coreos/gangplank/pkg/util"
)

// NewBuilder returns a Builder. NewBuilder determines what
// "Builder" to return.
func NewBuilder(ctx clustercontext.ClusterContext) (util.Builder, error) {
	inCluster := true
	if _, ok := os.LookupEnv(shared.LocalPodEnvVar); ok {
		log.Infof("EnvVar %s defined, using local pod mode", shared.LocalPodEnvVar)
		inCluster = false
	}

	ws, err := workspec.NewBuilder(ctx)
	if err == nil {
		return ws, nil
	}
	bc, err := bc.NewBuilder(ctx, &clustercontext.Cluster{InCluster: inCluster})
	if err == nil {
		return bc, nil
	}
	return nil, errors.ErrNoWorkFound
}
