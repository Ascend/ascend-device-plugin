// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import "sync/atomic"

// AtomicBool is an atomic Boolean.
type AtomicBool struct{ v uint32 }

// NewAtomicBool creates a AtomicBool.
func NewAtomicBool(initial bool) *AtomicBool {
	return &AtomicBool{v: boolToUint(initial)}
}

// Load atomically loads the Boolean.
func (b *AtomicBool) Load() bool {
	return atomic.LoadUint32(&b.v) == 1
}

// Store atomically stores the passed value.
func (b *AtomicBool) Store(new bool) {
	atomic.StoreUint32(&b.v, boolToUint(new))
}

func boolToUint(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
