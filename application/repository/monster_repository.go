package repository

import "RoyalWikiOverlay/domain"

type MonsterRepository interface {
	FindByLocation(locationName string) ([]domain.Monster, error)
	Save(monster domain.Monster) error
}
