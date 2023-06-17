package v1

import "time"

type PersonResource struct {
	Spec   PersonResourceSpec   `json:"spec"`
	Status PersonResourceStatus `json:"status"`
}

type PersonResourceSpec struct {
	FullName   string `json:"fullName"`
	KnownAs    string
	FamilyName string                     `json:"familyName,inline"`
	FamilyKey  string                     `json:"familyKey,omitempty"`
	Aliases    []string                   `json:"aliases,omitempty"`
	Children   []PersonReference          `json:"children,omitempty"`
	Friends    map[string]PersonReference `json:"friends,omitempty"`
	BirthDate  *time.Time                 `json:"birthDate,omitempty"`
}

type PersonResourceStatus struct {
	Age *int `json:"age,omitempty"`
}

type PersonReference struct {
	Name string `json:"name"`
}
