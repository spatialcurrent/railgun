// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"path/filepath"
)

// InferFormatAndCompression infers the format and compresion of the file.
//  - *.json => ("json", "") // JSON File
//  - *.json.bz2 => ("json", "bzip2") // bzip2-compressed JSON file
//  - *.json.gz => ("json", "gzip") // gzip-compressed JSON file
//  - *.json.sz => ("json", "snappy") // snappy-compressed JSON file
//  - *.jsonl => ("jsonl", "") // JSON Lines file
//  - *.jsonl.bz2 => ("jsonl", "bzip2") // bzip2-compressed JSON Lines file
//  - *.jsonl.gz => ("jsonl", "gzip") // gzip-compressed JSON Lines file
//  - *.jsonl.sz => ("jsonl", "snappy") // snappy-compressed JSON Lines file
//  - *.yaml => ("yaml", "") // YAML file
//  - *.yaml.bz2 => ("yaml", "bzip2") // bzip2-compressed YAML file
//  - *.yaml.gz => ("yaml", "gzip") // gzip-compressed YAML file
//  - *.yaml.sz => ("yaml", "snappy") // Snappy-compressed YAML file
func InferFormatAndCompression(p string) (string, string) {

	compression := ""

	ext := filepath.Ext(p)

	if len(ext) == 0 {
		return "", ""
	}

	if ext == ".enc" {
		p = p[:len(p)-4]
		ext = filepath.Ext(p)
		if len(ext) == 0 {
			return "", ""
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
	}

	if len(ext) == 0 {
		return "", ""
	}

	switch ext {
	case ".geojson":
		return "json", compression
	case ".json":
		return "json", compression
	case ".yaml":
		return "yaml", compression
	case ".yml":
		return "yaml", compression
	case ".properties":
		return "properties", compression
	}

	return "", compression
}
