package domain

type Monster struct {
	ID        int64
	Name      string
	Items     []Item
	Locations []Location
}

//TODO добавить колонку "тип" босс/элитка/обычный
