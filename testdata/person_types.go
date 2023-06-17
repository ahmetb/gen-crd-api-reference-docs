package v1

type PersonResource struct {
	Spec PersonResourceSpec `json:"spec"`
}

type PersonResourceSpec struct {
	FullName   string `json:"fullName"`
	KnownAs    string `json:"knownAs"`
	FamilyName string `json:"familyName,inline"`
	FamilyKey  string `json:"familyKey,omitempty"`
}
