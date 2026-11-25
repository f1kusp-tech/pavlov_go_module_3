package mocks

import (
	"testing"

	mock "github.com/stretchr/testify/mock"
)

type URLSaver struct {
	mock.Mock
}

func NewURLSaver(t testing.TB) *URLSaver {
	t.Helper()
	m := &URLSaver{}
	t.Cleanup(func() {
		m.AssertExpectations(t)
	})
	return m
}

func (m *URLSaver) Save(urlToSave, alias string) (string, error) {
	args := m.Called(urlToSave, alias)
	return args.String(0), args.Error(1)
}