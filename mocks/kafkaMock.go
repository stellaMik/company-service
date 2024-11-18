package mocks

import (
	"company-service/kafka"
	"github.com/stretchr/testify/mock"
)

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) ProduceEvent(event *kafka.EventMessage) error {
	args := m.Called(event)
	return args.Error(0)
}
func (m *MockKafkaProducer) Close() {
	m.Called()
}
