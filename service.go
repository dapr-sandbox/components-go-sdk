/*
Copyright 2022 The Dapr Authors
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

package dapr

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ErrSocketNotDefined is returned when the env variable `DAPR_COMPONENT_SOCKET_PATH` is not set
var ErrSocketNotDefined = errors.New("socket env `DAPR_COMPONENT_SOCKET_PATH` must be set")

const (
	unixSocketPathEnvVar = "DAPR_COMPONENT_SOCKET_PATH"
)

// makeAbortChan Generates a chan bool that automatically gets closed when the process
// receives a SIGINT or SIGTERM.
func makeAbortChan() chan struct{} {
	abortChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)

	go func() {
		<-sigChan
		close(abortChan)
	}()

	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	return abortChan
}

// Run starts the component server with the given options.
func Run(opts ...Option) error {
	socket, ok := os.LookupEnv(unixSocketPathEnvVar)
	if !ok {
		return ErrSocketNotDefined
	}

	// remove socket if it is already created.
	if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	svcLogger.Infof("using socket defined at '%s'", socket)

	lis, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}

	defer lis.Close()

	abort := makeAbortChan()
	go func() { // must close listener to abort while trying to accept connection
		<-abort
		lis.Close()
	}()

	svcOpts := &componentOpts{}
	for _, opt := range opts {
		opt(svcOpts)
	}

	server := grpc.NewServer()

	if err = svcOpts.apply(server); err != nil {
		return err
	}

	reflection.Register(server)
	return server.Serve(lis)
}

// MustRun same as run but panics on error
func MustRun(opts ...Option) {
	if err := Run(opts...); err != nil {
		panic(err)
	}
}
