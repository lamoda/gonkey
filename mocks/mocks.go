package mocks

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Mocks struct {
	mocks map[string]*ServiceMock
}

func New(mocks ...*ServiceMock) *Mocks {
	mocksMap := make(map[string]*ServiceMock, len(mocks))
	for _, v := range mocks {
		mocksMap[v.ServiceName] = v
	}
	return &Mocks{
		mocks: mocksMap,
	}
}

func NewNop(serviceNames ...string) *Mocks {
	mocksMap := make(map[string]*ServiceMock, len(serviceNames))
	for _, name := range serviceNames {
		mocksMap[name] = NewServiceMock(name, newDefinition("$", nil, &nopReply{}, callsNoConstraint))
	}
	return &Mocks{
		mocks: mocksMap,
	}
}

func (m *Mocks) ResetDefinitions() {
	for _, v := range m.mocks {
		v.ResetDefinition()
	}
}

func (m *Mocks) Start() error {
	for _, v := range m.mocks {
		err := v.StartServer()
		if err != nil {
			m.Shutdown()
			return err
		}
	}
	return nil
}

// Stops immediately, with no gracefully closing connections
func (m *Mocks) Shutdown() {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	_ = m.ShutdownContext(ctx)
}

func (m *Mocks) ShutdownContext(ctx context.Context) error {
	errs := make([]string, 0, len(m.mocks))
	for _, v := range m.mocks {
		if err := v.ShutdownServer(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", v.mock.path, err.Error()))
		}
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (m *Mocks) Service(serviceName string) *ServiceMock {
	mock, _ := m.mocks[serviceName]
	return mock
}

func (m *Mocks) ResetRunningContext() {
	for _, v := range m.mocks {
		v.ResetRunningContext()
	}
}

func (m *Mocks) EndRunningContext() []error {
	var errors []error
	for _, v := range m.mocks {
		errors = append(errors, v.EndRunningContext()...)
	}
	return errors
}
