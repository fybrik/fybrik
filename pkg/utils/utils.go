// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"

	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apiserver/pkg/storage/names"
)

const (
	StepNameHashLength       = 10
	hashPostfixLength        = 5
	S3MaxConformNameLength   = 63
	K8sMaxNameLength         = names.MaxGeneratedNameLength - 1 // We keep extra space for a "-"
	helmMaxConformNameLength = 53
)

// Intersection finds a common subset of two given sets of strings
func Intersection(set1, set2 []string) []string {
	res := []string{}
	for _, elem1 := range set1 {
		for _, elem2 := range set2 {
			if elem1 == elem2 {
				res = append(res, elem1)
				break
			}
		}
	}
	return res
}

func ListeningAddress(port int) string {
	address := fmt.Sprintf(":%d", port)
	if runtime.GOOS == "darwin" {
		address = "localhost" + address
	}
	return address
}

// StructToMap converts a struct to a map using JSON marshal
func StructToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

func HasString(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// Hash generates a name based on the unique identifier
func Hash(value string, hashLength int) string {
	data := sha512.Sum512([]byte(value))
	hashedStr := hex.EncodeToString(data[:])
	if hashLength >= len(hashedStr) {
		return hashedStr
	}
	return hashedStr[:hashLength]
}

// This function shortens a name to the maximum length given and uses rest of the string that is too long
// as hash that gets added to the valid name.
func ShortenedName(name string, maxLength, hashLength int) string {
	if len(name) > maxLength {
		// The new name is in the form prefix-suffix
		// The prefix is the prefix of the original name (so it's human readable)
		// The suffix is a deterministic hash of the suffix of the original name
		// Overall, the new name is deterministic given the original name
		cutOffIndex := maxLength - hashLength - 1
		prefix := name[:cutOffIndex]
		suffix := Hash(name[cutOffIndex:], hashLength)
		return prefix + "-" + suffix
	}
	return name
}

// Conforms a string to be a k8s compatible for an object name.
// The method first validates the name is a valid DNS subdomain, and hashes otherwise.
// In case the string is a valid DNS subdmain, the method shortens the name keeping a prefix and adds a hash of the suffix.
// The final output length is ActualMaxGeneratedNameLength.
func K8sConformName(name string, logger *zerolog.Logger) string {
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		logger.Info().Msg("Not according to k8s requirements: " + name + ", Hashing")
		hashLength := int(math.Min(K8sMaxNameLength, float64(len(name))))
		return Hash(name, hashLength)
	}
	return ShortenedName(name, K8sMaxNameLength, hashPostfixLength)
}

// Conforms a string to be a S3-compatible bucket name. Currently only shortens the name.
func S3ConformName(name string) string {
	return ShortenedName(name, S3MaxConformNameLength, hashPostfixLength)
}

// Conforms a string to be a Helm-compatible. Currently only shortens the name.
// Helm has stricter restrictions than K8s and restricts release names to 53 characters.
func HelmConformName(name string) string {
	return ShortenedName(name, helmMaxConformNameLength, hashPostfixLength)
}

// Given a path this function returns true if such path exists.
// Otherwise it returns false.
func IsPathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
