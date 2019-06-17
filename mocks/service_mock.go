package mocks

import (
	"context"
	"net"
	"net/http"
	"sync"
)

type ServiceMock struct {
	server            *http.Server
	listener          net.Listener
	mock              *definition
	defaultDefinition *definition
	sync.Mutex
	errors []error

	ServiceName string
}

func NewServiceMock(serviceName string, mock *definition) *ServiceMock {
	return &ServiceMock{
		mock:              mock,
		defaultDefinition: mock,
		ServiceName:       serviceName,
	}
}

func (m *ServiceMock) StartServer() error {
	addr := ":0" // all interfaces, random port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.listener = ln
	m.server = &http.Server{Addr: addr, Handler: m}
	go m.server.Serve(ln)
	return nil
}

func (m *ServiceMock) ShutdownServer() {
	m.server.Shutdown(context.TODO())
}

func (m *ServiceMock) ServerAddr() string {
	return m.listener.Addr().String()
}

func (m *ServiceMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()
	if m.mock != nil {
		errs := m.mock.Execute(w, r)
		for _, e := range errs {
			m.errors = append(m.errors, &Error{
				error:       e,
				ServiceName: m.ServiceName,
			})
		}
	}
}

func (m *ServiceMock) SetDefinition(newDefinition *definition) {
	m.Lock()
	defer m.Unlock()
	m.mock = newDefinition
}

func (m *ServiceMock) ResetDefinition() {
	m.mock = m.defaultDefinition
}

func (m *ServiceMock) ResetRunningContext() {
	m.errors = nil
	m.mock.ResetRunningContext()
}

func (m *ServiceMock) EndRunningContext() []error {
	errs := append(m.errors, m.mock.EndRunningContext()...)
	for i, e := range errs {
		errs[i] = &Error{
			error:       e,
			ServiceName: m.ServiceName,
		}
	}
	return errs
}
