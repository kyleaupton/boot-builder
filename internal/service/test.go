package service

import (
	"context"
)

type TestService struct{}

func NewTestService() *TestService { return &TestService{} }

func (s *TestService) TestPrivileged(ctx context.Context) (string, error) {
	return "test", nil
}
