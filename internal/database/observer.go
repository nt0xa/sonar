package database

import "github.com/russtone/sonar/internal/database/models"

type Observer interface {
	PayloadCreated(p models.Payload)
	PayloadDeleted(p models.Payload)
}

type DefaultObserver struct{}

var _ Observer = (*DefaultObserver)(nil)

func (o *DefaultObserver) PayloadCreated(p models.Payload) {}
func (o *DefaultObserver) PayloadDeleted(p models.Payload) {}
