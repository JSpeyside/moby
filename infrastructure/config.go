package infrastructure

import (
	"fmt"
	"strconv"
	"strings"
)

// Config holds the consts and config of the client.
type Config struct {
	Version         *Version
	RequiredCompose *Version
}

// Version stores a version of a product in an easy to compare way
type Version struct {
	Major int
	Minor int
	Point int
}

func NewVersion(major int, minor int, point int) *Version {
	return &Version{
		Major: major,
		Minor: minor,
		Point: point,
	}
}

func NewVersionFromStr(version string) (*Version, error) {
	matchSplit := strings.Split(version, ".")
	if len(matchSplit) != 3 {
		return nil, fmt.Errorf("Invalid version found %s", version)
	}
	major, err := strconv.Atoi(matchSplit[0])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(matchSplit[1])
	if err != nil {
		return nil, err
	}
	point, err := strconv.Atoi(matchSplit[2])
	if err != nil {
		return nil, err
	}
	return NewVersion(major, minor, point), nil
}

func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Point)
	return s
}

func (v *Version) Gte(other Version) bool {
	if other.Major >= v.Major ||
		other.Minor >= v.Minor ||
		other.Point >= v.Point {
		return true
	}
	return false
}

// LoadConfig returns a new config type with the latest config.
func LoadConfig() *Config {
	return &Config{
		Version:         NewVersion(0, 0, 1),
		RequiredCompose: NewVersion(1, 13, 0),
	}
}
