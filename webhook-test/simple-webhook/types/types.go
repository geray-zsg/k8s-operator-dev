package types

// WHParameters defines the structure for webhook parameters
type WHParameters struct {
	LabelsToCheck []string
	TLSKey        string
	TLSCert       string
	TLSPort       string
}

// Enable or disable specific features as needed
type ConfigEnabel struct {
	PodEnvInjectedHandleMutate    bool
	NamespaceLabelsHandleValidate bool
}
