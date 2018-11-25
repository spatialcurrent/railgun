package pipeline

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

var boundingBoxFilterNode = dfl.MustParseCompile("filter(@, '(@geometry?.coordinates != null) and (@geometry.coordinates[0] within $bbox[0] and $bbox[2]) and (@geometry.coordinates[1] within $bbox[1] and $bbox[3])')")

var limitNode = dfl.MustParseCompile("limit(@, $limit)")

var geoJsonNode = dfl.MustParseCompile("map(@, '@properties -= {`_tile_x`, `_tile_y`, `_tile_z`}') | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}")

type Pipeline struct {
	Nodes []dfl.Node
}

func (p *Pipeline) FilterBoundingBox() *Pipeline {
	return &Pipeline{
		Nodes: append(p.Nodes, boundingBoxFilterNode),
	}
}

func (p *Pipeline) FilterCustom(filterNode dfl.Node) *Pipeline {
	return &Pipeline{
		Nodes: append(p.Nodes, dfl.Function{Name: "filter", MultiOperator: &dfl.MultiOperator{Arguments: []dfl.Node{
			dfl.Attribute{Name: ""},
			dfl.Literal{Value: filterNode.Dfl(dfl.DefaultQuotes, false, 0)},
		}}}),
	}
}

func (p *Pipeline) Limit() *Pipeline {
	return &Pipeline{
		Nodes: append(p.Nodes, limitNode),
	}
}

func (p *Pipeline) GeoJSON() *Pipeline {
	return &Pipeline{
		Nodes: append(p.Nodes, geoJsonNode),
	}
}

func (p *Pipeline) Evaluate(vars map[string]interface{}, inputObject interface{}) (interface{}, error) {
	_, outputObject, err := dfl.Pipeline{Nodes: p.Nodes}.Evaluate(
		vars,
		inputObject,
		dfl.DefaultFunctionMap,
		dfl.DefaultQuotes)
	return outputObject, err
}

func New() *Pipeline {
	return &Pipeline{
		Nodes: make([]dfl.Node, 0),
	}
}
