// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import (
	"net"
	"sync"
)

// NewLimiter limit the concurrent connection of Listener
func NewLimiter(listener net.Listener) net.Listener {
	return &limiter{
		Listener:  listener,
		semaphore: make(chan struct{}, MaxGRPCConcurrentStreams),
		done:      make(chan struct{}),
	}
}

type limiter struct {
	net.Listener
	semaphore chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

type limitConn struct {
	net.Conn
	once     sync.Once
	callback func()
}

// acquire get semaphore. if concurrency limit is reached, it will hang up
func (l *limiter) acquire() bool {
	select {
	case <-l.done:
		return false
	case l.semaphore <- struct{}{}:
		return true
	}
}

// release the semaphore.
func (l *limiter) release() {
	<-l.semaphore
}

// Accept : override the implement of net.Listener interface
func (l *limiter) Accept() (net.Conn, error) {
	acquired := l.acquire()

	c, err := l.Listener.Accept()
	if err != nil {
		// accept got err, release semaphore immediately
		if acquired {
			l.release()
		}
		return nil, err
	}
	return &limitConn{Conn: c, callback: l.release}, nil
}

// Close : override the implement of net.Listener interface
func (l *limiter) Close() error {
	err := l.Listener.Close()

	// ensure close once
	l.closeOnce.Do(func() {
		close(l.done)
	})
	return err
}

func (l *limitConn) Close() error {
	err := l.Conn.Close()
	// when conn is done, release the semaphore
	l.once.Do(l.callback)
	return err
}
