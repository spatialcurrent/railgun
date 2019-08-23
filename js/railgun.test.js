// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

const { process, formats } = global.gss;

const testObject = {
  "a": "x",
  "b": "y",
  "c": "z"
};

const testArray = [
  {
    "a": "x",
    "b": "y",
    "c": "z"
  },
  {
    "b": "g",
    "c": "h",
    "d": "i"
  }
];

function log(str) {
  console.log(str.replace(/\n/g, "\\n").replace(/\t/g, "\\t").replace(/"/g, "\\\""));
}

describe('railgun', () => {

  it('checks the available formats', () => {
    expect(formats).toEqual(["bson", "csv", "go", "json", "jsonl", "properties", "tags", "toml", "tsv", "hcl", "hcl2", "yaml"]);
  });

});

describe('process', () => {

  it('convert an object from csv to json', () => {
    var { str, err } = convert("a,b,c\nx,y,z\n", "csv", "json", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("[{\"a\":\"x\",\"b\":\"y\",\"c\":\"z\"}]");
  });

  it('convert an object from csv to tsv', () => {
    var { str, err } = convert("a,b,c\nx,y,z\n", "csv", "tsv", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("a\tb\tc\nx\ty\tz\n");
  });

  it('convert an object from csv to yaml', () => {
    var { str, err } = convert("a,b,c\nx,y,z\n", "csv", "yaml", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("- a: x\n  b: \"y\"\n  c: z\n");
  });

  it('convert an object from json to bson and back', () => {
    var { str, err } = convert(JSON.stringify(testObject), "json", "bson");
    expect(err).toBeNull();
    // TODO: hex or base64 encode string so we can test
    //expect(Buffer.to(str, 'hex')).toEqual("a=x\nb=y\nc=z");
    var { str, err } = convert(str, "bson", "json");
    expect(err).toBeNull();
    expect(str).toEqual(JSON.stringify(testObject));
  });

  it('convert an object from json to properties and back', () => {
    var { str, err } = convert(JSON.stringify(testObject), "json", "properties", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("a=x\nb=y\nc=z");
    var { str, err } = convert(str, "properties", "json", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual(JSON.stringify(testObject));
  });

  it('convert an object from json to tags and back', () => {
    var { str, err } = convert(JSON.stringify(testObject), "json", "tags", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("a=x b=y c=z");
    var { str, err } = convert(str, "tags", "json", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual(JSON.stringify([testObject])); // to support streaming, deserializing tags returns an array
  });

  it('convert an object from json to toml and back', () => {
    var { str, err } = convert(JSON.stringify(testObject), "json", "toml", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("a = \"x\"\nb = \"y\"\nc = \"z\"\n");
    var { str, err } = convert(str, "toml", "json", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual(JSON.stringify(testObject));
  });

  it('convert an object from json to yaml and back', () => {
    var { str, err } = convert(JSON.stringify(testObject), "json", "yaml", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual("a: x\nb: \"y\"\nc: z\n");
    var { str, err } = convert(str, "yaml", "json", undefined, {"sorted": true});
    expect(err).toBeNull();
    expect(str).toEqual(JSON.stringify(testObject));
  });

});
