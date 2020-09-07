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
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"

	"github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	"github.com/Comcast/kuberhealthy/v2/pkg/kubeClient"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	buildVersion string = ""
	buildTime    string = ""

	awsSession *session.Session
	k8sClient  *kubernetes.Clientset

	awsRegionEnv = os.Getenv("AWS_REGION")
	awsRegion    string

	kubeConfigFile string = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	debugEnv       string = os.Getenv("DEBUG")
	debug          bool
	ctx            context.Context

	signalChan chan os.Signal
)

const (
	// Matching strings for kops bucket operations.
	regexpAWSRegion = `^[\w]{2}[-][\w]{4,9}[-][\d]$`

	// Default AWS region.
	defaultAWSRegion = "us-east-1"
)

func init() {

	// Parse AWS_REGION environment variable.
	awsRegion = defaultAWSRegion
	if len(awsRegionEnv) != 0 {
		awsRegion = awsRegionEnv
		ok, err := regexp.Match(regexpAWSRegion, []byte(awsRegion))
		if err != nil {
			log.Fatalln("Failed to parse AWS_REGION:", err.Error())
		}
		if !ok {
			log.Fatalln("Given AWS_REGION does not match AWS Region format.")
		}
	}

	// Create a signal chan for interrupts.
	signalChan = make(chan os.Signal, 2)

	// Create a context for this check.
	ctx = context.Background()

	var err error
	if len(debugEnv) != 0 {
		debug, err = strconv.ParseBool(debugEnv)
		if err != nil {
			log.Fatalln("Failed to parse DEBUG Environment variable:", err.Error())
		}
	}

	if debug {
		checkclient.Debug = debug
		log.SetLevel(log.DebugLevel)
	}

	// APP Build information
	log.Debugln("Application Version:", buildVersion)
	log.Debugln("Application Build Time:", buildTime)
}

func main() {
	var err error

	k8sClient, err = kubeClient.Create(kubeConfigFile)
	if err != nil {
		log.Fatalln("Unable to create kubernetes client")
		err = checkclient.ReportFailure([]string{err.Error()})
		os.Exit(1)
	}

	awsSession = createAWSSession()
	if awsSession == nil {
		err = fmt.Errorf("nil AWS session: %v", awsSession)
		err = checkclient.ReportFailure([]string{err.Error()})
		if err != nil {
			log.Println(err.Error())
		}
		os.Exit(1)
	}

	// Start listening for interrupts in the background.
	go listenForInterrupts()

	// Catch panics.
	var r interface{}
	defer func() {
		r = recover()
		if r != nil {
			log.Infoln("Recovered panic:", r)
			err = checkclient.ReportFailure([]string{r.(string)})
		}
	}()

	// Run the check.
	runCheck()
}

// listenForInterrupts watches the signal and done channels for termination.
func listenForInterrupts() {

	// Relay incoming OS interrupt signals to the signalChan.
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	sig := <-signalChan // This is a blocking operation -- the routine will stop here until there is something sent down the channel.
	log.Infoln("Received an interrupt signal from the signal channel.")
	log.Debugln("Signal received was:", sig.String())

	// Clean up pods here.
	log.Infoln("Shutting down.")

	os.Exit(0)
}

// reportErrorsToKuberhealthy reports the specified errors for this check run.
func reportErrorsToKuberhealthy(errs []string) {
	log.Errorln("Reporting errors to Kuberhealthy:", errs)
	reportToKuberhealthy(false, errs)
}

// reportOKToKuberhealthy reports that there were no errors on this check run to Kuberhealthy.
func reportOKToKuberhealthy() {
	log.Infoln("Reporting success to Kuberhealthy.")
	reportToKuberhealthy(true, []string{})
}

// reportToKuberhealthy reports the check status to Kuberhealthy.
func reportToKuberhealthy(ok bool, errs []string) {
	var err error
	if ok {
		err = checkclient.ReportSuccess()
		if err != nil {
			log.Fatalln("error reporting to kuberhealthy:", err.Error())
		}
		return
	}
	err = checkclient.ReportFailure(errs)
	if err != nil {
		log.Fatalln("error reporting to kuberhealthy:", err.Error())
	}
	return
}

// CustomRetryer wraps the SDK's built in DefaultRetryer adding additional
// custom features. Such as, no retry for 5xx status codes.
type CustomRetryer struct {
	client.DefaultRetryer
}

// ShouldRetry overrides the SDK's built in DefaultRetryer adding customization
// to not retry 5xx status codes.
func (r CustomRetryer) ShouldRetry(req *request.Request) bool {
	if req.HTTPResponse.StatusCode >= 500 {
		// Don't retry any 5xx status codes.
		return false
	}
	log.Debugln("Retrying AWS API call")

	// Fallback to SDK's built in retry rules
	return r.DefaultRetryer.ShouldRetry(req)
}
