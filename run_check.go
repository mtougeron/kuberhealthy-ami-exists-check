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
	"fmt"

	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func runCheck() {

	instanceIDs := getNodeInstanceIDs()

	awsEC2, _ := newEC2Client()

	var instanceResult InstanceAMIsResult
	select {
	case instanceResult = <-awsEC2.listEC2InstanceAMIs(instanceIDs):
		// Handle errors from listing AMIs.
		if instanceResult.Err != nil {
			log.Errorln("failed to list Instance AMIs:", instanceResult.Err.Error())
			err := fmt.Errorf("failed to list Instance AMIs: %w", instanceResult.Err)
			reportErrorsToKuberhealthy([]string{err.Error()})
			return
		}
		log.Infof("Retrieved Unique AWS Instance AMIs. (Total: %d)", len(instanceResult.InstanceAMIs))
	case <-ctx.Done():
		// If there is a context cancellation, exit the check.
		log.Infoln("Exiting check due to cancellation:", ctx.Err().Error())
		return
	}

	// Get a list of AMIs from AWS.
	var amiResult AMIResult
	select {
	case amiResult = <-awsEC2.listEC2Images(instanceResult.InstanceAMIs):
		// Handle errors from listing AMIs.
		if amiResult.Err != nil {
			log.Errorln("failed to list AMIs:", amiResult.Err.Error())
			err := fmt.Errorf("failed to list AMIs: %w", amiResult.Err)
			reportErrorsToKuberhealthy([]string{err.Error()})
			return
		}
		log.Infof("Retrieved AWS AMIs. (Total: %d)", len(amiResult.Images))
	case <-ctx.Done():
		// If there is a context cancellation, exit the check.
		log.Infoln("Exiting check due to cancellation:", ctx.Err().Error())
		return
	}

	if len(instanceResult.InstanceAMIs) == len(amiResult.Images) {
		reportOKToKuberhealthy()
	} else {
		reportErrorsToKuberhealthy([]string{"The number of AMIs found and unique Instance AMIs do not match."})
	}

}
