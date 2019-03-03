// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"path/filepath"
)

// SplitNameFormatCompression splits a filename into it's basename, format, and compression.
//  - *.json => ("*", "json", "") // JSON File
//  - *.json.bz2 => ("*", "json", "bzip2") // bzip2-compressed JSON file
//  - *.json.gz => ("*", "json", "gzip") // gzip-compressed JSON file
//  - *.json.sz => ("*", "json", "snappy") // snappy-compressed JSON file
//  - *.jsonl => ("*", "jsonl", "") // JSON Lines file
//  - *.jsonl.bz2 => ("*", "jsonl", "bzip2") // bzip2-compressed JSON Lines file
//  - *.jsonl.gz => ("*", "jsonl", "gzip") // gzip-compressed JSON Lines file
//  - *.jsonl.sz => ("*", "jsonl", "snappy") // snappy-compressed JSON Lines file
//  - *.yaml => ("*", "yaml", "") // YAML file
//  - *.yaml.bz2 => ("*", "yaml", "bzip2") // bzip2-compressed YAML file
//  - *.yaml.gz => ("*", "yaml", "gzip") // gzip-compressed YAML file
//  - *.yaml.sz => ("*", "yaml", "snappy") // Snappy-compressed YAML file
//  - *.hcl => ("*", "hcl", "") // HCL file
//  - *.hcl.bz2 => ("*", "hcl", "bzip2") // bzip2-compressed HCL file
//  - *.hcl.gz => ("*", "hcl", "gzip") // gzip-compressed HCL file
//  - *.hcl.sz => ("*", "hcl", "snappy") // Snappy-compressed HCL file
func SplitNameFormatCompression(p string) (string, string, string) {

	compression := ""

	ext := filepath.Ext(p)

	if len(ext) == 0 {
		return p, "", ""
	}

	if ext == ".enc" {
		p = p[:len(p)-4]
		ext = filepath.Ext(p)
		if len(ext) == 0 {
			return p, "", ""
		}
	}

	if ext == ".gz" {
		compression = "gzip"
		p = p[:len(p)-3]
		ext = filepath.Ext(p)
	} else if ext == ".sz" {
		compression = "snappy"
		p = p[:len(p)-3]
		ext = filepath.Ext(p)
	} else if ext == ".bz2" {
		compression = "bzip2"
		p = p[:len(p)-4]
		ext = filepath.Ext(p)
	} else if ext == ".zip" {
		compression = "zip"
		p = p[:len(p)-4]
		ext = filepath.Ext(p)
	}

	if len(ext) == 0 {
		return p, "", compression
	} else {
		p = p[:len(p)-len(ext)]
	}

	switch ext {
	case ".csv":
		return p, "csv", compression
	case ".tsv":
		return p, "tsv", compression
	case ".geojson":
		return p, "json", compression
	case ".bson":
		return p, "bson", compression
	case ".json":
		return p, "json", compression
	case ".jsonl":
		return p, "jsonl", compression
	case ".geojsonl":
		return p, "jsonl", compression
	case ".html":
		return p, "html", compression
	case ".yaml":
		return p, "yaml", compression
	case ".yml":
		return p, "yaml", compression
	case ".properties", ".props", ".prop":
		return p, "properties", compression
	case ".tf":
		return p, "hcl", compression
	case ".hcl":
		return p, "hcl", compression
	case ".toml":
		return p, "toml", compression
	}

	return p, "", compression

}
