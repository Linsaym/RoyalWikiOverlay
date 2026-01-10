package repository

import "RoyalWikiOverlay/domain"

type LocationRepository interface {
	GetByID(id int64) (*domain.Location, error)
	GetByName(name string) (*domain.Location, error)
	List() ([]domain.Location, error)
	Upsert(location domain.Location) error
}
