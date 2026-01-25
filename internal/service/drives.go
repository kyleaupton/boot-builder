package service

import (
	"github.com/kyleaupton/flashit/internal/drives"
	"context"
)

type DrivesService struct{}

func NewDrivesService() *DrivesService { return &DrivesService{} }

func (s *DrivesService) ListDrives(ctx context.Context) ([]drives.Drive, error) {
	return drives.ListRemovable(ctx)
}
