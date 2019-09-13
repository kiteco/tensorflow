/*
Copyright 2016 The TensorFlow Authors. All Rights Reserved.

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

package tensorflow

// #include "tensorflow/c/c_api.h"
import "C"

import (
	"runtime"
	"sync"
)

type code C.TF_Code

// status holds error information returned by TensorFlow. We convert all
// TF statuses to Go errors.
type status struct {
	m sync.Mutex
	c *C.TF_Status
}

func newStatus() *status {
	s := &status{c: C.TF_NewStatus()}
	runtime.SetFinalizer(s, (*status).finalizer)
	return s
}

func (s *status) finalizer() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.c != nil {
		C.TF_DeleteStatus(s.c)
	}
	s.c = nil
}

func (s *status) Delete() {
	s.finalizer()
}

func (s *status) Code() code {
	s.m.Lock()
	defer s.m.Unlock()
	return s.codeLocked()
}

// codeLocked returns the code without locking - only to be used when the caller already holds the lock
func (s *status) codeLocked() code {
	return code(C.TF_GetCode(s.c))
}

func (s *status) String() string {
	s.m.Lock()
	defer s.m.Unlock()
	return s.stringLocked()
}

func (s *status) stringLocked() string {
	return C.GoString(C.TF_Message(s.c))
}

// Err converts the status to a Go error and returns nil if the status is OK.
func (s *status) Err() error {
	s.m.Lock()
	defer s.m.Unlock()
	// Note that we call codeLocked here vs Code() because we are already holding the lock
	if s == nil || s.codeLocked() == C.TF_OK {
		return nil
	}

	return statusError(s.stringLocked())
}

// statusError is distinct from status because it fulfills the error interface.
// status itself may have a TF_OK code and is not always considered an error.
//
// TODO(jhseu): Make public, rename to Error, and provide a way for users to
// check status codes.
type statusError string

func (s statusError) Error() string {
	return string(s)
}
