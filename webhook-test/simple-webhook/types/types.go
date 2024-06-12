package types

// WHParameters defines the structure for webhook parameters
type WHParameters struct {
	// server parameters
	TLSKey     string
	TLSCert    string
	TLSPort    string
	HealthPort string

	//	handlers parameters
	// LabelsToCheck    []string
	DeploymentPrefix string
}

// Enable or disable specific features as needed
type ConfigEnabel struct {
	PodEnvInjectedHandleMutate    bool
	NamespaceLabelsHandleValidate bool
	CheckDeploymentPrefix         bool
}

type ConfigHandlersParameters struct {
	//	handlers parameters
	LabelsToCheck    []string
	DeploymentPrefix string
}
