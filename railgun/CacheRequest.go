package railgun

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

type CacheRequest struct {
	Key string
	Hit bool
}

func (cr CacheRequest) String() string {
	str := "cache"
	if cr.Hit {
		str += " hit"
	} else {
		str += " miss"
	}
	str += " for key " + cr.Key
	return str
}

func (cr CacheRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"key": cr.Key,
		"hit": cr.Hit,
	}
}

func (cr CacheRequest) Serialize(format string) (string, error) {
	return gss.SerializeString(cr.Map(), format, []string{}, -1)
}
