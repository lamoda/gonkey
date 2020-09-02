package mocks

import (
	"errors"
	"strings"
	"sync"
)

type Mocks struct {
	mocks    map[string]*ServiceMock
	shutDown sync.Once
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
	var (
		wg      = sync.WaitGroup{}
		errChan = make(chan error, len(m.mocks))
	)
	for _, v := range m.mocks {
		wg.Add(1)
		go func(v *ServiceMock) {
			defer wg.Done()
			if internalErr := v.StartServer(); internalErr != nil {
				m.Shutdown()
				errChan <- internalErr
			}
		}(v)
	}
	wg.Wait()
	close(errChan)
	errList := make([]string, 0, len(errChan))
	for err := range errChan {
		errList = append(errList, err.Error())
	}
	if len(errList) != 0 {
		return errors.New(strings.Join(errList, "; "))
	}
	return nil
}

func (m *Mocks) Shutdown() {
	m.shutDown.Do(func() {
		wg := sync.WaitGroup{}
		for _, v := range m.mocks {
			wg.Add(1)
			go func(v *ServiceMock) {
				defer wg.Done()
				v.ShutdownServer()
			}(v)
		}
		wg.Wait()
	})
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
