package shared

const (
	// OcpStructTag is the struct tag used to read in
	// OCPBuilder from envvars
	OcpStructTag = "envVar"

	// DefaultContextDir is the default path to use for a build
	DefaultContextDir = "/srv"

	// SecretLabelName is the label to search for secrets to automatically use
	SecretLabelName = "coreos-assembler.coreos.com/secret"

	// CosaMetaJSON is the meta.json file
	CosaMetaJSON = "meta.json"

	// CosaBuildsJSON is the COSA build.json file name
	CosaBuildsJSON = "builds.json"

	// SourceBin stores binary input
	SourceBin = "source.bin"

	// SourceSubPath is used when extracting binary inputs
	SourceSubPath = "source"

	// MinioRegion is a "fake" region
	MinioRegion = "darkarts-1"

	PodBuildLabel      = "gangplank.coreos-assembler.coreos.com"
	PodBuildAnnotation = PodBuildLabel + "%s"
	PodBuildRunnerTag  = "cosa-podBuild-runner"

	KvmLabel       = "devices.kubevirt.io/kvm"
	LocalPodEnvVar = "COSA_FORCE_NO_CLUSTER"
)
