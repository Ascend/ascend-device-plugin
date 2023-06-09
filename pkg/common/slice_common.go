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

type int32Tool struct {
}

type int64Tool struct {
}

// Int32Tool slice for int32 tool
var Int32Tool int32Tool

// Int64Tool slice for int64 tool
var Int64Tool int64Tool

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
	for len(sources) > 0 {
		index := i.Index(sources, target)
		if index == -1 {
			return sources
		}
		sources = append(sources[:index], sources[index+1:]...)
	}
	return sources
}

// Index slice for int6 search the index with target
func (i int64Tool) Index(sources []int64, target int64) int {
	for i, source := range sources {
		if source == target {
			return i
		}
	}
	return -1
}
