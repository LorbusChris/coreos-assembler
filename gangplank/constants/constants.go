package constants

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

	// podmanCaps are the specific permissions we needed to run a podman
	// pod. This is a privileged pod.
	podmanCaps = []string{
		"CAP_DAC_READ_SEARCH",
		"CAP_LINUX_IMMUTABLE",
		"CAP_NET_BROADCAST",
		"CAP_NET_ADMIN",
		"CAP_IPC_LOCK",
		"CAP_IPC_OWNER",
		"CAP_SYS_MODULE",
		"CAP_SYS_RAWIO",
		"CAP_SYS_PTRACE",
		"CAP_SYS_PACCT",
		"CAP_SYS_ADMIN",
		"CAP_SYS_BOOT",
		"CAP_SYS_NICE",
		"CAP_SYS_RESOURCE",
		"CAP_SYS_TIME",
		"CAP_SYS_TTY_CONFIG",
		"CAP_LEASE",
		"CAP_AUDIT_CONTROL",
		"CAP_MAC_OVERRIDE",
		"CAP_MAC_ADMIN",
		"CAP_SYSLOG",
		"CAP_WAKE_ALARM",
		"CAP_BLOCK_SUSPEND",
		"CAP_AUDIT_READ",
	}

	// MinioRegion is a "fake" region
	MinioRegion = "darkarts-1"
)
