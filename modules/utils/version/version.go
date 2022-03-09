package version

import "github.com/hashicorp/go-version"

func Greater(v1, v2 string) (bool, error) {
	sv1, err := version.NewSemver(v1)
	if err != nil {
		return false, err
	}
	sv2, err := version.NewSemver(v2)
	if err != nil {
		return false, err
	}
	return sv1.GreaterThan(sv2), nil
}

func GreaterOrEqual(v1, v2 string) (bool, error) {
	sv1, err := version.NewSemver(v1)
	if err != nil {
		return false, err
	}
	sv2, err := version.NewSemver(v2)
	if err != nil {
		return false, err
	}
	return sv1.GreaterThanOrEqual(sv2), nil
}
