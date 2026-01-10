package repository

import "RoyalWikiOverlay/domain"

type LocationRepository interface {
	FindByName(name string) (*domain.Location, error)
	Save(location domain.Location) error
}
