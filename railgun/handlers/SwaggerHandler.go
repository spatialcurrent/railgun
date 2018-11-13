// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-swagger-structs/swagger"
	"github.com/spatialcurrent/railgun/railgun/core"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type SwaggerHandler struct {
	*BaseHandler
}

func (h *SwaggerHandler) getRequiredProperties(t reflect.Type) []string {
	properties := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if str, ok := f.Tag.Lookup("rest"); ok && str != "" && str != "-" {
			if r, ok := f.Tag.Lookup("required"); ok && (r == "true" || r == "t" || r == "1" || r == "y" || r == "yes") {
				if strings.Contains(str, ",") {
					properties = append(properties, strings.TrimSpace(strings.SplitN(str, ",", 2)[0]))
				} else {
					properties = append(properties, strings.TrimSpace(str))
				}
			}
		}
	}
	return properties
}

func (h *SwaggerHandler) getProperties(t reflect.Type) map[string]swagger.Property {
	properties := map[string]swagger.Property{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if str, ok := f.Tag.Lookup("rest"); ok && str != "" && str != "-" {
			if strings.Contains(str, ",") {
				properties[strings.TrimSpace(strings.SplitN(str, ",", 2)[0])] = swagger.Property{Type: "string"}
			} else {
				properties[strings.TrimSpace(str)] = swagger.Property{Type: "string"}
			}
		}
	}
	return properties
}

func (h *SwaggerHandler) BuildPaths(singular string, plural string, basepath string, t reflect.Type) map[string]swagger.Path {
	m := map[string]swagger.Path{}
	tags := []string{strings.Title(plural)}
	ext := swagger.Parameter{
		Name:        "ext",
		Type:        "string",
		Description: "File extension",
		In:          "path",
		Required:    true,
		Default:     "json",
		Enumeration: []string{"bson", "json", "yaml", "toml"},
	}

	m[fmt.Sprintf("/%s.{ext}", basepath)] = swagger.Path{
		Get: swagger.Operation{
			Description: fmt.Sprintf("list %s on Railgun Server", plural),
			Tags:        tags,
			Produces: []string{
				"application/json",
				"text/yaml",
				"application/ubjson",
				"application/toml",
			},
			Parameters: []swagger.Parameter{ext},
			Responses: map[string]swagger.Response{
				"200": swagger.Response{
					Description: "Success",
				},
			},
		},
		Post: swagger.Operation{
			Description: fmt.Sprintf("add %s to Railgun Server", singular),
			Tags:        tags,
			Consumes: []string{
				"application/json",
				"text/yaml",
				"application/ubjson",
				"application/toml",
			},
			Produces: []string{
				"application/json",
				"text/yaml",
				"application/ubjson",
				"application/toml",
			},
			Parameters: []swagger.Parameter{
				swagger.Parameter{
					Name:        singular,
					Type:        "",
					Description: fmt.Sprintf("the %s to add to the Railgun Server", singular),
					In:          "body",
					Required:    true,
					Schema: &swagger.Schema{
						Ref: fmt.Sprintf("#/definitions/%s", strings.Title(strings.Replace(singular, " ", "", -1))),
					},
				},
				ext,
			},
			Responses: map[string]swagger.Response{
				"200": swagger.Response{
					Description: "Success",
				},
				"400": swagger.Response{
					Description: fmt.Sprintf("Bad request. %s with provided name already exists.", strings.Title(singular)),
				},
			},
		},
	}
	m[fmt.Sprintf("/%s/{name}.{ext}", basepath)] = swagger.Path{
		Get: swagger.Operation{
			Description: fmt.Sprintf("get %s from Railgun Server", plural),
			Tags:        tags,
			Produces: []string{
				"application/json",
				"text/yaml",
				"application/ubjson",
				"application/toml",
			},
			Parameters: []swagger.Parameter{
				swagger.Parameter{
					Name:        "name",
					Type:        "string",
					Description: fmt.Sprintf("the name of the %s on the Railgun Server", singular),
					In:          "path",
					Required:    true,
				},
				ext,
			},
			Responses: map[string]swagger.Response{
				"200": swagger.Response{
					Description: "Success",
				},
				"404": swagger.Response{
					Description: fmt.Sprintf("Not found. %s with provided name was not found.", strings.Title(singular)),
				},
			},
		},
		Delete: swagger.Operation{
			Description: fmt.Sprintf("delete %s from Railgun Server", plural),
			Tags:        tags,
			Produces: []string{
				"application/json",
				"text/yaml",
				"application/ubjson",
				"application/toml",
			},
			Parameters: []swagger.Parameter{
				swagger.Parameter{
					Name:        "name",
					Type:        "string",
					Description: fmt.Sprintf("the name of the %s on the Railgun Server", singular),
					In:          "path",
					Required:    true,
				},
				ext,
			},
			Responses: map[string]swagger.Response{
				"200": swagger.Response{
					Description: "OK",
				},
				"400": swagger.Response{
					Description: fmt.Sprintf("Bad request.  Could not delete %s with provided name.", singular),
				},
				"404": swagger.Response{
					Description: fmt.Sprintf("Not found. %s with provided name was not found.", strings.Title(singular)),
				},
				"500": swagger.Response{
					Description: fmt.Sprintf("Server error while deleting %s with provided name.", singular),
				},
			},
		},
	}
	return m
}

func (h *SwaggerHandler) BuildDefinitions() map[string]swagger.Definition {
	definitions := map[string]swagger.Definition{}
	definitions["Credentials"] = swagger.Definition{
		Type:     "object",
		Required: []string{"username", "password"},
		Properties: map[string]swagger.Property{
			"username": swagger.Property{Type: "string"},
			"password": swagger.Property{Type: "string"},
		},
	}
	for name, t := range core.CoreTypes {
		definitions[strings.Title(name)] = swagger.Definition{
			Type:       "object",
			Required:   h.getRequiredProperties(t),
			Properties: h.getProperties(t),
		}
	}
	return definitions
}

func (h *SwaggerHandler) BuildSwaggerDocument() (swagger.Document, error) {

	location, err := url.Parse(h.Viper.GetString("http-location"))
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
			Enumeration: []string{"bson", "json", "yaml", "toml"},
		},
		"name": swagger.Parameter{
			Name:        "name",
			Type:        "string",
			Description: "The name",
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

	paths := map[string]swagger.Path{
		"/authenticate.{ext}": swagger.Path{
			Get: swagger.Operation{
				Description: "Authenticate",
				Tags:        []string{"Security"},
				Parameters: []swagger.Parameter{
					swagger.Parameter{
						Name:        "credentials",
						Type:        "",
						Description: "the login credentials",
						In:          "body",
						Required:    true,
						Schema: &swagger.Schema{
							Ref: fmt.Sprintf("#/definitions/%s", "Credentials"),
						},
					},
					params["ext"],
				},
				Responses: map[string]swagger.Response{
					"200": swagger.Response{
						Description: "OK",
					},
				},
			},
		},
		"/swagger.{ext}": swagger.Path{
			Get: swagger.Operation{
				Description: "Railgun Swagger Document",
				Tags:        []string{"Swagger"},
				Parameters:  []swagger.Parameter{params["ext"]},
				Responses: map[string]swagger.Response{
					"200": swagger.Response{
						Description: "OK",
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
						Description: "OK",
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
						Description: "OK",
					},
				},
			},
		},
		"/services/exec.{ext}": swagger.Path{
			Post: swagger.Operation{
				Description: "execute a service on the Railgun Server with the given job input",
				Tags:        []string{"Services"},
				Consumes: []string{
					"application/json",
					"text/yaml",
					"application/ubjson",
					"application/toml",
				},
				Produces: []string{
					"application/json",
					"text/yaml",
					"application/ubjson",
					"application/toml",
				},
				Parameters: []swagger.Parameter{
					swagger.Parameter{
						Name:        "job",
						Type:        "",
						Description: fmt.Sprintf("the %s to execute on the Railgun Server", "job"),
						In:          "body",
						Required:    true,
						Schema: &swagger.Schema{
							Ref: fmt.Sprintf("#/definitions/%s", "Job"),
						},
					},
					params["ext"],
				},
				Responses: map[string]swagger.Response{
					"200": swagger.Response{
						Description: "OK",
					},
					"404": swagger.Response{
						Description: fmt.Sprintf("Not found. %s with provided name was not found.", "service"),
					},
				},
			},
		},
		"/jobs/{name}/exec.{ext}": swagger.Path{
			Post: swagger.Operation{
				Description: "execute a job for a service on the Railgun Server",
				Tags:        []string{"Jobs"},
				Consumes: []string{
					"application/json",
					"text/yaml",
					"application/ubjson",
					"application/toml",
				},
				Produces: []string{
					"application/json",
					"text/yaml",
					"application/ubjson",
					"application/toml",
				},
				Parameters: []swagger.Parameter{
					swagger.Parameter{
						Name:        "name",
						Type:        "string",
						Description: fmt.Sprintf("the name of the %s to execute on the Railgun Server", "job"),
						In:          "path",
						Required:    true,
					},
					params["ext"],
				},
				Responses: map[string]swagger.Response{
					"200": swagger.Response{
						Description: "OK",
					},
					"404": swagger.Response{
						Description: fmt.Sprintf("Not found. %s with provided name was not found.", "service"),
					},
				},
			},
		},
		"/workflows/{name}/exec.{ext}": swagger.Path{
			Post: swagger.Operation{
				Description: "execute a workflow for a service on the Railgun Server",
				Tags:        []string{"Workflows"},
				Consumes: []string{
					"application/json",
					"text/yaml",
					"application/ubjson",
					"application/toml",
				},
				Produces: []string{
					"application/json",
					"text/yaml",
					"application/ubjson",
					"application/toml",
				},
				Parameters: []swagger.Parameter{
					swagger.Parameter{
						Name:        "name",
						Type:        "string",
						Description: fmt.Sprintf("the name of the %s to execute on the Railgun Server", "workflow"),
						In:          "path",
						Required:    true,
					},
					params["ext"],
				},
				Responses: map[string]swagger.Response{
					"200": swagger.Response{
						Description: "OK",
					},
					"404": swagger.Response{
						Description: fmt.Sprintf("Not found. %s with provided name was not found.", "workflow"),
					},
				},
			},
		},
		"/layers/{name}/data/tiles/{z}/{x}/{y}.{ext}": swagger.Path{
			Get: swagger.Operation{
				Description: "Get GeoJSON tile of features filtered by a DFL expression.",
				Tags:        []string{"Layers"},
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
		"/layers/{name}/mask/tiles/{z}/{x}/{y}.{ext}": swagger.Path{
			Get: swagger.Operation{
				Description: "Get mask tile of features filtered by a DFL expression.",
				Tags:        []string{"Layers"},
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
	}

	for k, v := range h.BuildPaths("workspace", "workspaces", "workspaces", core.WorkspaceType) {
		paths[k] = v
	}

	for k, v := range h.BuildPaths("data store", "data stores", "datastores", core.DataStoreType) {
		paths[k] = v
	}

	for k, v := range h.BuildPaths("layer", "layers", "layers", core.LayerType) {
		paths[k] = v
	}

	for k, v := range h.BuildPaths("process", "processes", "processes", core.ProcessType) {
		paths[k] = v
	}

	for k, v := range h.BuildPaths("service", "services", "services", core.ServiceType) {
		paths[k] = v
	}

	for k, v := range h.BuildPaths("job", "jobs", "jobs", core.JobType) {
		paths[k] = v
	}

	for k, v := range h.BuildPaths("workflow", "workflows", "workflows", core.WorkflowType) {
		paths[k] = v
	}

	var contact *swagger.Contact
	swaggerContactName := h.Viper.GetString("swagger-contact-name")
	swaggerContactEmail := h.Viper.GetString("swagger-contact-email")
	swaggerContactUrl := h.Viper.GetString("swagger-contact-url")
	if len(swaggerContactName) > 0 || len(swaggerContactEmail) > 0 || len(swaggerContactUrl) > 0 {
		contact = &swagger.Contact{
			Name:  h.Viper.GetString("swagger-contact-name"),
			Email: h.Viper.GetString("swagger-contact-email"),
			Url:   h.Viper.GetString("swagger-contact-url"),
		}
	}

	doc := swagger.Document{
		Version:  "2.0",
		Schemes:  h.Viper.GetStringSlice("http-schemes"),
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
			Contact:        contact,
			License: &swagger.License{
				Name: "MIT",
				Url:  "https://github.com/spatialcurrent/railgun/blob/master/LICENSE",
			},
		},
		Paths:       paths,
		Definitions: h.BuildDefinitions(),
	}

	return doc, nil
}

func (h *SwaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	swaggerDocument, err := h.BuildSwaggerDocument()
	if err != nil {
		h.Messages <- err
		return
	}

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(swaggerDocument, format, []string{}, -1)
	if err != nil {
		h.Messages <- err
		return
	}
	w.Write(b)

}
