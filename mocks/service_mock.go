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
	return m.StartServerWithAddr("localhost:0") // loopback, random port
}

func (m *ServiceMock) StartServerWithAddr(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.listener = ln
	m.server = &http.Server{Addr: addr, Handler: m}
	go func() {
		_ = m.server.Serve(ln)
	}()
	return nil
}

func (m *ServiceMock) ShutdownServer(ctx context.Context) error {
	err := m.server.Shutdown(ctx)
	m.listener = nil
	m.server = nil
	return err
}

func (m *ServiceMock) ServerAddr() string {
	if m.listener == nil {
		panic("mock server " + m.ServiceName + " is not started")
	}
	return m.listener.Addr().String()
}

func (m *ServiceMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	if m.mock != nil {
		errs := m.mock.Execute(w, r)
		m.errors = append(m.errors, errs...)
	}
}

func (m *ServiceMock) SetDefinition(newDefinition *definition) {
	m.Lock()
	defer m.Unlock()
	m.mock = newDefinition
}

func (m *ServiceMock) ResetDefinition() {
	m.Lock()
	defer m.Unlock()
	m.mock = m.defaultDefinition
}

func (m *ServiceMock) ResetRunningContext() {
	m.Lock()
	defer m.Unlock()
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
