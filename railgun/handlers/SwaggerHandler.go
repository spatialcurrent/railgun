package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-swagger-structs/swagger"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
	"net/url"
)

type SwaggerHandler struct {
	*BaseHandler
}

func (h *SwaggerHandler) BuildSwaggerDocument() (swagger.Document, error) {

	v := h.Viper

	location, err := url.Parse(v.GetString("http-location"))
	if err != nil {
		return swagger.Document{}, err
	}

	params := map[string]swagger.Parameter{
		"ext": swagger.Parameter{
			Name:        "ext",
			Type:        "string",
			Description: "File extension",
			In:          "path",
			Required:    true,
			Default:     "json",
			Enumeration: []string{"json", "yaml"},
		},
		"name": swagger.Parameter{
			Name:        "name",
			Type:        "string",
			Description: "The collection name",
			In:          "path",
			Required:    true,
			Default:     "",
		},
		"z": swagger.Parameter{
			Name:        "z",
			Type:        "integer",
			Description: "The tile zoom level (1 - 18)",
			In:          "path",
			Required:    true,
			Default:     0,
			Minimum:     aws.Int(0),
			Maximum:     aws.Int(18),
		},
		"x": swagger.Parameter{
			Name:        "x",
			Type:        "integer",
			Description: "The tile X coordinate",
			In:          "path",
			Required:    true,
			Default:     0,
		},
		"y": swagger.Parameter{
			Name:        "y",
			Type:        "integer",
			Description: "The tile y coordinate",
			In:          "path",
			Required:    true,
			Default:     0,
		},
		"dfl": swagger.Parameter{
			Name:        "dfl",
			Type:        "string",
			Description: "The DFL expression",
			In:          "query",
			Required:    false,
			Default:     "",
		},
		"limit": swagger.Parameter{
			Name:        "limit",
			Type:        "integer",
			Description: "Limit the number of results to this maximum count",
			In:          "query",
			Required:    false,
			Default:     0,
			Minimum:     aws.Int(0),
		},
	}

	doc := swagger.Document{
		Version:  "2.0",
		Schemes:  v.GetStringSlice("http-schemes"),
		BasePath: location.Path,
		Host:     location.Host,
		External: &swagger.External{
			Description: "A simple and fast data processing tool",
			Url:         "https://github.com/spatialcurrent/railgun",
		},
		Info: &swagger.Info{
			Version:        "1.0.0",
			Title:          "Railgun",
			Description:    "A simple and fast data processing tool",
			TermsOfService: "https://spatialcurrent.io/terms-of-service/",
			Contact: &swagger.Contact{
				Email: v.GetString("swagger-contact-email"),
			},
			License: &swagger.License{
				Name: "MIT",
				Url:  "https://github.com/spatialcurrent/railgun/blob/master/LICENSE",
			},
		},
		Paths: map[string]swagger.Path{
			"/swagger.{ext}": swagger.Path{
				Get: swagger.Operation{
					Description: "Railgun Swagger Document",
					Tags:        []string{"Swagger"},
					Parameters:  []swagger.Parameter{params["ext"]},
					Responses: map[string]swagger.Response{
						"200": swagger.Response{
							Description: "Success",
						},
					},
				},
			},
			"/dfl/functions.{ext}": swagger.Path{
				Get: swagger.Operation{
					Description: "DFL Functions",
					Tags:        []string{"DFL"},
					Parameters:  []swagger.Parameter{params["ext"]},
					Responses: map[string]swagger.Response{
						"200": swagger.Response{
							Description: "Success",
						},
					},
				},
			},
			"/gss/formats.{ext}": swagger.Path{
				Get: swagger.Operation{
					Description: "go-simple-serializer (GSS) Formats",
					Tags:        []string{"GSS"},
					Parameters:  []swagger.Parameter{params["ext"]},
					Responses: map[string]swagger.Response{
						"200": swagger.Response{
							Description: "Success",
						},
					},
				},
			},
			"/collections/{name}/data/tiles/{z}/{x}/{y}.{ext}": swagger.Path{
				Get: swagger.Operation{
					Description: "Get GeoJSON tile of features filtered by a DFL expression.",
					Tags:        []string{"Data"},
					Parameters: []swagger.Parameter{
						params["name"],
						params["z"],
						params["x"],
						params["y"],
						swagger.Parameter{
							Name:        "ext",
							Type:        "string",
							Description: "File extension",
							In:          "path",
							Required:    true,
							Default:     "json",
							Enumeration: []string{"json", "jsonl", "yaml", "geojson", "geojsonl"},
						},
						params["dfl"],
						swagger.Parameter{
							Name:        "buffer",
							Type:        "integer",
							Description: "The number of tiles to buffer by.",
							In:          "query",
							Required:    false,
							Default:     0,
							Minimum:     aws.Int(0),
						},
						params["limit"],
					},
					Responses: map[string]swagger.Response{
						"200": swagger.Response{
							Description: "Success",
						},
					},
				},
			},
			"/collections/{name}/mask/tiles/{z}/{x}/{y}.{ext}": swagger.Path{
				Get: swagger.Operation{
					Description: "Get mask tile of features filtered by a DFL expression.",
					Tags:        []string{"Data"},
					Parameters: []swagger.Parameter{
						params["name"],
						params["z"],
						params["x"],
						params["y"],
						swagger.Parameter{
							Name:        "ext",
							Type:        "string",
							Description: "File extension",
							In:          "path",
							Required:    true,
							Default:     "json",
							Enumeration: []string{"png", "jpg", "jpeg", "json", "yaml"},
						},
						params["dfl"],
						params["limit"],
						swagger.Parameter{
							Name:        "theshold",
							Type:        "integer",
							Description: "The minimum threshold for the cell to be considered in the region.",
							In:          "query",
							Required:    false,
							Default:     0,
							Minimum:     aws.Int(0),
						},
						swagger.Parameter{
							Name:        "zoom",
							Type:        "integer",
							Description: "The mask zoom level (1 - 18)",
							In:          "query",
							Required:    true,
							Default:     16,
							Minimum:     aws.Int(0),
							Maximum:     aws.Int(18),
						},
						swagger.Parameter{
							Name:        "alpha",
							Type:        "integer",
							Description: "The mask alpha level (0 - 255)",
							In:          "query",
							Required:    true,
							Default:     255,
							Minimum:     aws.Int(0),
							Maximum:     aws.Int(255),
						},
					},
					Responses: map[string]swagger.Response{
						"200": swagger.Response{
							Description: "Success",
						},
					},
				},
			},
		},
	}

	return doc, nil
}

func (h *SwaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	swaggerDocument, err := h.BuildSwaggerDocument()
	if err != nil {
		h.Messages <- err
		return
	}

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(swaggerDocument, format, []string{}, -1)
	if err != nil {
		h.Messages <- err
		return
	}
	w.Write(b)

}
