package repository

import "RoyalWikiOverlay/domain"

type MonsterRepository interface {
	GetByID(id int64) (*domain.Monster, error)
	ListByLocation(locationID int64) ([]domain.Monster, error)
	Upsert(monster domain.Monster) error
}
