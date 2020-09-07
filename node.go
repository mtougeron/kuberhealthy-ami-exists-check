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
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func parseInstanceID(providerID string) *string {
	parts := strings.Split(providerID, "/")
	return aws.String(parts[len(parts)-1])
}

func getNodeInstanceIDs() []*string {
	nodes, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Errorln("Error getting nodes")
		log.Debugln(err.Error())
		reportErrorsToKuberhealthy([]string{"Could not get nodes: " + err.Error()})
		os.Exit(1)
	}

	log.Debugf("Found %d nodes", len(nodes.Items))
	if len(nodes.Items) == 0 {
		log.Errorf("There were no nodes found")
		reportErrorsToKuberhealthy([]string{"There were no nodes found"})
		os.Exit(1)
	}

	var instanceIDs []*string
	for _, node := range nodes.Items {
		instanceIDs = append(instanceIDs, parseInstanceID(node.Spec.ProviderID))
	}

	return instanceIDs
}
