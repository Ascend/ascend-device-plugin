/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package common a series of common function
package common

import (
	"strconv"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
)

type int32Tool struct {
}

type int64Tool struct {
}

type stringTool struct {
}

// Int32Tool slice for int32 tool
var Int32Tool int32Tool

// Int64Tool slice for int64 tool
var Int64Tool int64Tool

// StringTool slice for string tool
var StringTool stringTool

// Contains slice for int32 contains
func (i int32Tool) Contains(sources []int32, target int32) bool {
	for _, sourceNum := range sources {
		if sourceNum == target {
			return true
		}
	}
	return false
}

// SameElement slice for int64 has same element with others slice
func (i int64Tool) SameElement(sources, targets []int64) bool {
	for _, source := range sources {
		for _, target := range targets {
			if source == target {
				return true
			}
		}
	}
	return false
}

// Remove slice for int64 remove target
func (i int64Tool) Remove(sources []int64, target int64) []int64 {
	if len(sources) == 0 {
		return sources
	}
	index := i.Index(sources, target)
	if index == -1 {
		return sources
	}
	return i.Remove(append(sources[:index], sources[index+1:]...), target)
}

// Index slice for int64 search the index with target
func (i int64Tool) Index(sources []int64, target int64) int {
	for i, source := range sources {
		if source == target {
			return i
		}
	}
	return -1
}

// ToHexString slice for int64 to Hex string
func (i int64Tool) ToHexString(sources []int64) string {
	var target string
	for i, source := range sources {
		if i == 0 {
			target = strconv.FormatInt(source, Hex)
			continue
		}
		target = target + "," + strconv.FormatInt(source, Hex)
	}
	return target
}

// Index slice for string search the index with target
func (s stringTool) Index(sources []string, target string) int {
	for i, source := range sources {
		if source == target {
			return i
		}
	}
	return -1
}

// SameElement string slice has same element with others slice
func (s stringTool) SameElement(sources, targets []string) bool {
	for _, source := range sources {
		for _, target := range targets {
			if source == target {
				return true
			}
		}
	}
	return false
}

// HexStringToInt hex string slice to int64 slice
func (s stringTool) HexStringToInt(sources []string) []int64 {
	intSlice := make([]int64, 0, len(sources))
	for _, source := range sources {
		num, err := strconv.ParseInt(source, Hex, 0)
		if err != nil {
			hwlog.RunLog.Errorf("parse hex int failed , string: %s", source)
			return nil
		}
		intSlice = append(intSlice, num)
	}
	return intSlice
}
