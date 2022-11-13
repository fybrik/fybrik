// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package serde

import (
	"encoding/json"

	"github.com/mohae/deepcopy"
)

type Properties struct {
	Items map[string]interface{} `json:"-"`
}

func (in *Properties) DeepCopyInto(out *Properties) {
	// TODO: missing type assertion
	cp, _ := deepcopy.Copy(in).(*Properties)
	*out = *cp
}

func (in *Properties) DeepCopy() *Properties {
	if in == nil {
		return nil
	}
	out := new(Properties)
	in.DeepCopyInto(out)
	return out
}

func (in *Properties) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &in.Items)
	return err
}

func (in *Properties) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.Items)
}

// replace a templated string with the given interface inside a map
func replaceTemplateInMap(m map[string]interface{}, template string, value interface{}) {
	for key, val := range m {
		m[key] = replaceTemplate(val, template, value)
	}
}

// return an interface after a templated field has been replaced with the given value
func replaceTemplate(inter interface{}, template string, value interface{}) interface{} {
	switch inter := inter.(type) {
	case string:
		if inter == template {
			return value
		}
	case map[string]interface{}:
		replaceTemplateInMap(inter, template, value)
	}
	return inter
}

// ReplaceTemplateWithValue replaces a templated string with the given value
func (in *Properties) ReplaceTemplateWithValue(template string, value interface{}) {
	replaceTemplateInMap(in.Items, template, value)
}

func matchPatternInMap(actual, templated map[string]interface{}, pattern string) interface{} {
	for templateKey, templateValue := range templated {
		actualValue := actual[templateKey]
		if templateValue == pattern {
			return actualValue
		}
		if actualValue == nil {
			continue
		}
		val := matchPattern(actualValue, templateValue, pattern)
		if val != nil {
			return val
		}
	}
	return nil
}

func matchPattern(actual, templated interface{}, pattern string) interface{} {
	actMap, ok := actual.(map[string]interface{})
	if !ok {
		return nil
	}
	tMap, ok := templated.(map[string]interface{})
	if !ok {
		return nil
	}
	return matchPatternInMap(actMap, tMap, pattern)
}

// MatchPattern returns the value of the templated property that matches the given pattern, or nil if not found
func (in *Properties) MatchPattern(templatedProp Properties, pattern string) interface{} {
	return matchPatternInMap(in.Items, templatedProp.Items, pattern)
}
