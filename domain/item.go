package domain

type ItemType string

const (
	ItemTypeCard      ItemType = "card"
	ItemTypeEquipment ItemType = "equipment"
	ItemTypeOther     ItemType = "other"
)

type Item struct {
	ID      int64
	Name    string
	Type    ItemType
	Price   int64
	WikiURL string
}
