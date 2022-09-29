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
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	// ErrNoComponentsRegistered is returned when none components was registered.
	ErrNoComponentsRegistered = errors.New("none components was registered")
)

const (
	unixSocketFolderPathEnvVar = "DAPR_COMPONENT_SOCKET_FOLDER"
	defaultSocketFolder        = "/tmp/dapr-components-sockets"
)

// makeAbortChan Generates a chan bool that automatically gets closed when the process
// receives a SIGINT or SIGTERM.
func makeAbortChan(done chan struct{}) chan struct{} {
	abortChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)

	go func() {
		select {
		case <-sigChan:
			close(abortChan)
		case <-done:
			close(abortChan)

		}
	}()

	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	return abortChan
}

func runComponent(socket string, opts *componentsOpts, abortChan chan struct{}) error {
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

	server := grpc.NewServer()

	if err = opts.apply(server); err != nil {
		return err
	}
	go func() {
		<-abortChan
		lis.Close()
	}()

	reflection.Register(server)
	return server.Serve(lis)
}

// Run starts the component server.
func Run() error {
	socketFolder, ok := os.LookupEnv(unixSocketFolderPathEnvVar)
	if !ok {
		socketFolder = defaultSocketFolder
	}
	if len(factories) == 0 {
		return ErrNoComponentsRegistered
	}
	done := make(chan struct{})
	abort := makeAbortChan(done)

	for component, opts := range factories {
		socket := filepath.Join(socketFolder, component+".sock")
		go func(opts *componentsOpts) {
			err := runComponent(socket, opts, abort)
			if err != nil {
				svcLogger.Errorf("aborting due to an error %v", err)
				done <- struct{}{}
			}
		}(opts)
	}

	select {
	case <-done:
		return nil
	case <-abort:
		return nil
	}
}

// MustRun same as run but panics on error
func MustRun() {
	if err := Run(); err != nil {
		panic(err)
	}
}
