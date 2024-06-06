package types

// WHParameters defines the structure for webhook parameters
type WHParameters struct {
	LabelsToCheck []string
	TLSKey        string
	TLSCert       string
}
