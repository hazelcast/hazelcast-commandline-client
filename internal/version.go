package internal

import (
	"fmt"
	"strconv"
	"strings"
)

const UnknownVersion = "UNKNOWN"
const CustomBuildSuffix = "CUSTOMBUILD"

// being initialized at compile-time.
var (
	GitCommit      string
	Version        = UnknownVersion
	IsCheckVersion = "disabled"
)

// CheckVersion checks whether left OP right condition holds.
func CheckVersion(left, operator, right string) bool {
	switch operator {
	case "=":
		return compareVersions(left, right) == 0
	case "!=":
		return compareVersions(left, right) != 0
	case ">":
		return compareVersions(left, right) > 0
	case ">=":
		return compareVersions(left, right) >= 0
	case "<":
		return compareVersions(left, right) < 0
	case "<=":
		return compareVersions(left, right) <= 0
	case "~":
		return looselyEqual(left, right)
	default:
		panic(fmt.Errorf(`unexpected test skip operator "%s" to compare versions`, operator))
	}
}

func compareVersions(left, right string) int {
	var leftHasSuffix, rightHasSuffix bool
	leftHasSuffix, left = stripSuffix(left)
	rightHasSuffix, right = stripSuffix(right)
	leftNums := strings.Split(left, ".")
	rightNums := strings.Split(right, ".")
	// make rightNums and leftNums the same length by filling the shorter one with zeros.
	if len(rightNums) < len(leftNums) {
		rightNums = equalizeVersions(rightNums, leftNums)
	} else if len(leftNums) < len(rightNums) {
		leftNums = equalizeVersions(leftNums, rightNums)
	}
	for i := 0; i < len(rightNums); i++ {
		r := mustAtoi(rightNums[i])
		l := mustAtoi(leftNums[i])
		if r < l {
			return 1
		}
		if r > l {
			return -1
		}
	}
	// if the version number has a suffix, then it is prior to the non-suffixed one.
	// See: https://semver.org/#spec-item-9
	if leftHasSuffix {
		if rightHasSuffix {
			return 0
		}
		return -1
	}
	if rightHasSuffix {
		return 1
	}
	return 0
}

// looselyEqual uses right version for the precision
func looselyEqual(left, right string) bool {
	_, left = stripSuffix(left)
	_, right = stripSuffix(right)
	leftNums := strings.Split(left, ".")
	rightNums := strings.Split(right, ".")
	minNums := len(rightNums)
	if len(leftNums) < minNums {
		// fill left num with zeros
		leftNums = equalizeVersions(leftNums, rightNums)
	}
	for i := 0; i < minNums; i++ {
		r := mustAtoi(rightNums[i])
		l := mustAtoi(leftNums[i])
		if r != l {
			return false
		}
	}
	return true
}

// stripSuffix checks and removes suffix, e.g., "-beta1" from the version.
func stripSuffix(version string) (hasSuffix bool, newVersion string) {
	newVersion = version
	hasSuffix = strings.Contains(version, "-")
	if hasSuffix {
		newVersion = strings.SplitN(newVersion, "-", 2)[0]
	}
	return
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("could not parse int %s: %w", s, err))
	}
	return n
}

// equalizeVersions makes minV and maxV the same length by filling minV with zeros and returns the new minV
func equalizeVersions(minV, maxV []string) []string {
	var i int
	newMinV := make([]string, len(maxV))
	for i = 0; i < len(minV); i++ {
		newMinV[i] = minV[i]
	}
	for ; i < len(maxV); i++ {
		newMinV[i] = "0"
	}
	return newMinV
}
