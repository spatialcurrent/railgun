package railgun

import (
	"github.com/patrickmn/go-cache"
)

type Collection struct {
	Name        string       `json:"name"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	DataStore   DataStore    `json:"-"`
	Cache       *cache.Cache `json:"-"`
}
