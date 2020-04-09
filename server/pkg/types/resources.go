package types

// ResourceGroup : Resource Profile group
type ResourceGroup struct {
	ID               uint64
	Name             string
	ResourceProfiles []*ResourceProfile
}

// ResourceProfile : Resource Plans
type ResourceProfile struct {
	ID        uint64
	Name      string
	Disabled  bool
	Subtitle  string
	ShortDesc string
	LongDesc  string
	CPU       string
	GPU       string
	GPURam    string
	RAM       string
	Disk      string
}

// IsValidResourceProfile : Is valid resource profile?
func IsValidResourceProfile(p *ResourceProfile) bool {
	if p == nil {
		return false
	}
	if p.Name == nullString {
		return false
	}

	if p.CPU == nullString {
		return false
	}

	if p.RAM == nullString {
		return false
	}

	return true
}

// ResourceProfiles : List of resource plans
type ResourceProfiles []ResourceProfile

// ContainerImage : Container images
type ContainerImage struct {
	ID          uint64
	Name        string
	Disabled    bool
	RegistryURL string
	DescText    string
	DescHTML    string
}

// ContainerImages : List of container images
type ContainerImages []ContainerImage

// PythonLib : Name of the libraries that need to be installed
type PythonLib struct {
	Name     string
	DescText string
	DescHTML string
}

// PythonLibs : List of python libs
type PythonLibs []PythonLib

// EnvVar : List of OS environment vars
type EnvVar struct {
	Name  string
	Value string
}
