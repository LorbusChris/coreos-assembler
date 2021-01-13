package util

import (
	"runtime"

	"github.com/coreos/gangplank/pkg/clustercontext"
)

// Builder is the manual/unbounded Pod Build interface.
// It uses a build.openshift.io/v1 Build interface
// to use the exact same code path between the two.
type Builder interface {
	Exec(ctx clustercontext.ClusterContext) error
}

// BuilderArch converts the GOARCH to the build arch.
// In other words, it translates amd64 to x86_64.
func BuilderArch() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	return arch
}
