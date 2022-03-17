// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
)

const letters = "abcdefghijklmnopqrstuvwxyz"

type sortByLength []string

// Len implements Len of sort.Interface
func (s sortByLength) Len() int {
	return len(s)
}

// Swap implements Swap of sort.Interface
func (s sortByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements Less of sort.Interface
func (s sortByLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}

func getLongest(toFind []string) []string {
	// We sort it by length, descending
	sort.Sort(sortByLength(toFind))
	longest := []string{toFind[0]}

	// In case we have more than one element in toFind...
	if len(toFind) > 1 {
		for _, str := range toFind[1:] {
			if len(str) < len(longest[0]) {
				break
			}
			longest = append(longest, str)
		}
	}
	fmt.Println(longest)
	return longest
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
// Source: https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func GenerateRandomString(n int) (string, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// Source: https://stackoverflow.com/questions/26153441/generate-crypto-random-integer-beetwen-min-max-values
func GenerateRandomNumber(min, max int64) (int64, error) {
	// calculate the max we will be using
	bg := big.NewInt(max - min)

	// get big.Int between 0 and bg
	// in this case 0 to 20
	n, err := rand.Int(rand.Reader, bg)
	if err != nil {
		return -1, errors.WithMessage(err, "Error during random number generation in GenerateRandomNumber")
	}

	// add n to min to support the passed in range
	return n.Int64() + min, nil
}

func GenerateUniqueAssetName(namespace, namePrefix string, log *zerolog.Logger, client kclient.Client) (string, error) {
	var result v1alpha1.AssetList
	var randomStringLength = 1
	var uniqueAssetName = ""
	err := client.List(context.Background(), &result, kclient.InNamespace(namespace))
	if err == nil {
		listOfCandidates := make([]string, 0)
		for i := 0; i < len(result.Items); i++ {
			if strings.Contains(result.Items[i].Name, namePrefix) {
				listOfCandidates = append(listOfCandidates, result.Items[i].Name)
			}
		}
		log.Info().Msg("listOfCandidates : " + strings.Join(listOfCandidates, "|"))
		var randomStr string
		randomStr, err = GenerateRandomString(randomStringLength)
		if err == nil {
			if len(listOfCandidates) > 0 {
				longestArr := getLongest(listOfCandidates)
				const delimiter = "|"
				log.Info().Msg("longestArr : " + strings.Join(longestArr, delimiter))
				randIdx, err1 := GenerateRandomNumber(0, int64(len(longestArr)))
				if err1 != nil {
					return "", err1
				}
				log.Info().Msg("randIdx : " + fmt.Sprint(randIdx))
				uniqueAssetName = longestArr[randIdx] + randomStr
			} else {
				// no asset with the given prefix
				uniqueAssetName = namePrefix + randomStr
			}
			log.Info().Msg("uniqueAssetName generated : " + uniqueAssetName)
		} else {
			log.Info().Msg("Error during GenerateRandomString: " + err.Error())
		}
	} else {
		log.Info().Msg("Error during list operation: " + err.Error())
	}
	return uniqueAssetName, err
}
