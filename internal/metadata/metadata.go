/*
 *
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package metadata contains functions to set and get metadata from addresses.
//
// This package is experimental.
package metadata

import (
	"errors"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

type mdKeyType string

const mdKey = mdKeyType("grpc.internal.address.metadata")

type mdValue metadata.MD

func (m mdValue) Equal(o interface{}) bool {
	om, ok := o.(mdValue)
	if !ok {
		return false
	}
	if len(m) != len(om) {
		return false
	}
	for k, v := range m {
		ov := om[k]
		if len(ov) != len(v) {
			return false
		}
		for i, ve := range v {
			if ov[i] != ve {
				return false
			}
		}
	}
	return true
}

// Get returns the metadata of addr.
func Get(addr resolver.Address) metadata.MD {
	attrs := addr.Attributes
	if attrs == nil {
		return nil
	}
	md, _ := attrs.Value(mdKey).(mdValue)
	return metadata.MD(md)
}

// Set sets (overrides) the metadata in addr.
//
// When a SubConn is created with this address, the RPCs sent on it will all
// have this metadata.
func Set(addr resolver.Address, md metadata.MD) resolver.Address {
	addr.Attributes = addr.Attributes.WithValue(mdKey, mdValue(md))
	return addr
}

// Validate returns an error if the input md contains invalid keys or values.
//
// if header name not presudo-header, there are check items:
// - header names contain one or more characters from this set [0-9 a-z _ - .]
// - if the header-name ends with a "-bin" suffix, the header-value could contain an arbitrary octet sequence. So no real validation required here.
// - if header-name does not end with a "-bin" suffix, header-value should only contain one or more characters from the set ( %x20-%x7E ) which includes space and printable ASCII.
func Validate(md metadata.MD) error {
	for k, vals := range md {
		// presudo-header will be ignored
		if k[0] == ':' {
			continue
		}
		// check key, for i that saving a conversion if not using for range
		for i := 0; i < len(k); i++ {
			r := k[i]
			if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '.' && r != '-' && r != '_' {
				return errors.New("header key contains not 0-9a-z-_. characters")
			}
		}
		if strings.HasSuffix(k, "-bin") {
			continue
		}
		// check value
		for _, val := range vals {
			if hasNotPrintable(val) {
				return errors.New("header val contains not printable ASCII characters")
			}
		}
	}
	return nil
}

// hasNotPrintable return true if msg contains any characters which are not in %x20-%x7E
func hasNotPrintable(msg string) bool {
	// for i that saving a conversion if not using for range
	for i := 0; i < len(msg); i++ {
		if msg[i] < 0x20 || msg[i] > 0x7E {
			return true
		}
	}
	return false
}
