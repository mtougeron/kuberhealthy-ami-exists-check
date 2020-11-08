// Licensed to Michael Tougeron <github@e.tougeron.com> under
// one or more contributor license agreements. See the LICENSE
// file distributed with this work for additional information
// regarding copyright ownership.
// Michael Tougeron <github@e.tougeron.com> licenses this file
// to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// NOTE: This check leverages heavily the work done in the Kuberhealthy
// ami-check https://github.com/Comcast/kuberhealthy/tree/master/cmd/ami-check

package main

import (
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func Test_parseInstanceID(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		want       string
	}{
		{
			name:       "full providerID",
			providerID: "aws:///us-east-1d/i-089747b9fac6ab469",
			want:       "i-089747b9fac6ab469",
		},
		{
			name:       "partial providerID",
			providerID: "i-abc123",
			want:       "i-abc123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInstanceID(tt.providerID); got != tt.want {
				t.Errorf("parseInstanceID() = %v, want %v", got, tt.want)
			}
		})
	}
}
