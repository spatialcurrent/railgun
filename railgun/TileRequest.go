package railgun

import (
	"fmt"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"net/http"
)

type TileRequest struct {
	Header        http.Header
	Collection    string
	Tile          Tile
	Bbox          []float64
	Source        string
	Expression    string
	Features      int
	OutsideExtent bool
}

func (tr TileRequest) String() string {
	str := "requested tile " + tr.Tile.String() + " for collection " + tr.Collection + " from " + tr.Source
	if len(tr.Expression) > 0 {
		str += " with filter " + tr.Expression
	}
	str += " has " + fmt.Sprint(tr.Features) + " features within bounding box " + fmt.Sprint(tr.Bbox)
	return str
}

func (tr TileRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"collection": map[string]interface{}{
			"name": tr.Collection,
		},
		"bbox":          tr.Bbox,
		"outsideExtent": tr.OutsideExtent,
		"tile":          tr.Tile.Map(),
		"source": map[string]interface{}{
			"uri": tr.Source,
		},
		"http": map[string]interface{}{
			"header": tr.Header,
		},
		"dfl": map[string]interface{}{
			"expression": tr.Expression,
		},
		"results": map[string]interface{}{
			"features": tr.Features,
		},
	}
}

func (tr TileRequest) Serialize(format string) (string, error) {
	return gss.SerializeString(tr.Map(), format, []string{}, -1)
}
