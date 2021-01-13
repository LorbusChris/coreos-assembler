package pod

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	buildapiv1 "github.com/openshift/api/build/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/coreos/gangplank/pkg/buildv1/bc"
	"github.com/coreos/gangplank/pkg/clustercontext"
	"github.com/coreos/gangplank/pkg/shared"
	"github.com/coreos/gangplank/pkg/spec"
	"github.com/coreos/gangplank/pkg/util"
)

/*
	The Pod Builder is an interface for running Gangplank as part
	of a CI system (i.e. Jenkins) while benefiting from the BuildConfig
	niceties.

	Gangplank refers to non-buildconfig builders as "unbound"; they are not
	bound to a buildconfig and therefore run as part of some other procress such
	as a bare-pod, CI system, or CLI run. When run unbounded a mock OpenShift
	build.openshift.io/v1 object is created; this ensures that the same execution
	mode between all modes of running Gangplank.

	While it does not require running as a BuildConfig, it does require that the
	running pod exposes a service account with:
	- secret access
	- the ability to create pods
*/

type podBuild struct {
	apibuild *buildapiv1.Build
	bc       *bc.BuildConfig
	js       *spec.JobSpec

	clusterCtx clustercontext.ClusterContext
	pod        *v1.Pod

	hostname         string
	image            string
	ipaddr           string
	jobSpecFile      string
	projectNamespace string
	serviceAccount   string
	workDir          string
}

var (
	// cli is a Builder (and a poor one at that too...)
	// While the Pod Builder is a Builder, we treat it seperately.
	_ = util.Builder(&podBuild{})
)

// Exec start the unbounded build.
func (pb *podBuild) Exec(ctx clustercontext.ClusterContext) error {
	log.Info("Executing unbounded builder")
	return pb.bc.Exec(ctx)
}

// NewBuilder returns a pod builder ready for execution.
func NewBuilder(ctx clustercontext.ClusterContext, image, serviceAccount, jsF, workDir string) (util.Builder, error) {
	// Directly inject the jobspec
	js, err := spec.JobSpecFromFile(jsF)
	if jsF != "" && err != nil {
		return nil, fmt.Errorf("failed to read in jobspec from %s: %w", jsF, err)
	}
	if js.Recipe.GitURL == "" {
		return nil, fmt.Errorf("JobSpec %q does inclue a Git Recipe", jsF)
	}

	pb := &podBuild{
		clusterCtx:     ctx,
		image:          image,
		jobSpecFile:    jsF,
		js:             &js,
		serviceAccount: serviceAccount,
		workDir:        workDir,
	}

	if err := pb.setInCluster(); err != nil {
		return nil, fmt.Errorf("failed setting in-cluster options: %v", err)
	}

	// Generate the build.openshift.io/v1 object
	if err := pb.generateAPIBuild(); err != nil {
		return nil, fmt.Errorf("failed to generate api build: %v", err)
	}
	pbb, err := pb.encodeAPIBuild()
	if err != nil {
		return nil, fmt.Errorf("failed to encode apibuild: %v", err)
	}

	// Create the BuildConfig object
	os.Setenv("BUILD", pbb)
	bc, err := bc.NewBuilder(ctx, &clustercontext.Cluster{})
	if err != nil {
		return nil, err
	}
	bc.JobSpec = js
	bc.JobSpecFile = jsF

	pb.bc = bc
	return pb, nil
}

// setInCluster does the nessasary setup for unbounded builder running as
// an in-cluster build.
func (pb *podBuild) setInCluster() error {
	c, err := clustercontext.GetCluster(pb.clusterCtx)
	if err == nil && (c.Podman || !c.InCluster) {
		return nil
	}
	if err != nil {
		return err
	}

	// Dig deep and query find out what Kubernetes thinks this pod
	// Discover where this running
	hostname, ok := os.LookupEnv("HOSTNAME")
	if !ok {
		return errors.New("Unable to find hostname")
	}
	pb.hostname = hostname

	// Open the Kubernetes Client
	ac, pn, err := clustercontext.K8sInClusterClient()
	if err != nil {
		return fmt.Errorf("failed create a kubernetes client: %w", err)
	}
	pb.projectNamespace = pn

	myIP, err := clustercontext.GetPodIP(ac, pn, hostname)
	if err != nil {
		return fmt.Errorf("failed to query my hostname: %w", err)
	}
	pb.ipaddr = myIP

	// Discover where this running
	myPod, err := ac.CoreV1().Pods(pn).Get(hostname, metav1.GetOptions{})
	if err != nil {
		return err
	}
	pb.pod = myPod

	// Find the running pod this is running on. The controller pod should be
	// have "cosa" or "coreos-assembler" in the image name, otherwise the
	// image should be explicitly defined.
	var myContainer *v1.Container = nil
	for _, k := range myPod.Spec.Containers {
		lk := strings.ToLower(k.Image)
		for _, x := range []string{"cosa", "coreos-assembler"} {
			if strings.Contains(lk, x) {
				myContainer = &k
				break
			}
		}
	}

	// Allow both the service account and the image to be overriden.
	if pb.serviceAccount == "" {
		pb.serviceAccount = myPod.Spec.ServiceAccountName
	}
	if pb.image == "" {
		pb.image = myContainer.Image
	}
	if pb.serviceAccount == "" || pb.image == "" {
		return errors.New("serviceAccount and image must be defined by running pod or via overrides")
	}
	return nil
}

// generateAPIBuild creates a "mock" buildconfig.openshift.io/v1 Kubernetes
// object that is consumed by `bc.go`.
func (pb *podBuild) generateAPIBuild() error {
	// Create just _enough_ of the OpenShift BuildConfig spec
	// Create a "ci" build.openshift.io/v1 specification.
	podBuildNumber := time.Now().Format("20060102150405")
	a := buildapiv1.Build{}

	// Create annotations
	a.Annotations = map[string]string{
		// ciRunnerTag is tested for to determine if this is
		// a buildconfig or a faked one
		shared.PodBuildRunnerTag:                     "true",
		fmt.Sprintf(shared.PodBuildAnnotation, "IP"): pb.ipaddr,
		// Required Labels
		buildapiv1.BuildConfigAnnotation: "cosa",
		buildapiv1.BuildNumberAnnotation: podBuildNumber,
	}

	// Create basic labels
	a.Labels = map[string]string{
		shared.PodBuildLabel: podBuildNumber,
	}

	// Populate the Spec
	a.Spec = buildapiv1.BuildSpec{}
	a.Spec.ServiceAccount = pb.serviceAccount
	a.Spec.Strategy = buildapiv1.BuildStrategy{}
	a.Spec.Strategy.CustomStrategy = new(buildapiv1.CustomBuildStrategy)
	a.Spec.Strategy.CustomStrategy.From = v1.ObjectReference{
		Name: pb.image,
	}
	a.Spec.Source = buildapiv1.BuildSource{
		ContextDir: pb.workDir,
		Git: &buildapiv1.GitBuildSource{
			Ref: pb.js.Recipe.GitRef,
			URI: pb.js.Recipe.GitURL,
		},
	}

	pb.apibuild = &a
	return nil
}

// encodeAPIBuilder the ci buildapiv1 object to a JSON object.
// JSON is the messaginging interface for Kubernetes.
func (pb *podBuild) encodeAPIBuild() (string, error) {
	if pb.apibuild == nil {
		return "", errors.New("apibuild is not defined yet")
	}
	aW := bytes.NewBuffer([]byte{})
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, bc.BuildScheme, bc.BuildScheme)
	if err := s.Encode(pb.apibuild, aW); err != nil {
		return "", err
	}
	d, err := ioutil.ReadAll(aW)
	if err != nil {
		return "", err
	}

	return string(d), nil
}
