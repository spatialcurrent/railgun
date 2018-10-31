package railgun

type Collection struct {
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DataStore   DataStore `json:"-"`
	Cache       *Cache    `json:"-"`
}
