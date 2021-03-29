// Copyright 2020 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const codeTemplate string = `/* Code generated by codegen/main.go. DO NOT EDIT. */

use std::collections::HashMap;

use nanoserde::DeJson;


#[derive(Debug, Clone)]
pub enum Authentication {
  Basic {
    username: String,
    password: String
  },
  Bearer {
    token: String
  }
}

#[derive(Debug, Clone, PartialEq, Copy)]
pub enum Method {
    Post, Get, Put, Delete
}

#[derive(Debug, Clone)]
pub struct RestRequest<Response> {
  pub authentication: Authentication,
  pub urlpath: String,
  pub query_params: String,
  pub body: String,
  pub method: Method,
  _marker: std::marker::PhantomData<Response>
}

trait ToRestString {
    fn to_string(&self) -> String;
}

{{- range $defname, $definition := .Definitions }}
{{- $classname := $defname | title }}

/// {{ $definition.Description | stripNewlines }}
#[derive(Debug, DeJson, Default)]
#[nserde(default)]
pub struct {{ $classname }} {
    {{- range $propname, $property := $definition.Properties }}
    {{- $fieldname := $propname | snakeCase }}
    {{- $attrDataName := $propname | snakeCase }}

    {{- if eq $property.Type "integer" }}
    pub {{ $fieldname }}: i32,

    {{- else if eq $property.Type "number" }}
    pub {{ $fieldname }}: f32,

    {{- else if eq $property.Type "boolean" }}
    pub {{ $fieldname }}: bool,

    {{- else if eq $property.Type "string" }}
    pub {{ $fieldname }}: String,

    {{- else if eq $property.Type "array" }}
	{{- if eq $property.Items.Type "string" }}
    pub {{ $fieldname }}: Vec<String>,
	{{- else if eq $property.Items.Type "integer" }}
    pub {{ $fieldname }}: Vec<i32>,
	{{- else if eq $property.Items.Type "number" }}
    pub {{ $fieldname }}: Vec<f32>,
	{{- else if eq $property.Items.Type "boolean" }}
    pub {{ $fieldname }}: Vec<bool>,
	{{- else}}
    pub {{ $fieldname }}: Vec<{{ $property.Items.Ref | cleanRef }}>,
	{{- end }}
    {{- else if eq $property.Type "object"}}
	{{- if eq $property.AdditionalProperties.Type "string"}}
    pub {{ $fieldname }}: HashMap<String, String>,
	{{- else if eq $property.Items.Type "integer"}}
    pub {{ $fieldname }}: HashMap<String, i32>,
       {{- else if eq $property.Items.Type "number"}}
    pub {{ $fieldname }}: HashMap<String, f32>,
	{{- else if eq $property.Items.Type "boolean"}}
    pub {{ $fieldname }}: HashMap<string, bool>,
	{{- else}}
    pub {{ $fieldname }}: HashMap<string, {{$property.AdditionalProperties | cleanRef}}>,
	{{- end}}
    {{- else }}
    pub {{ $fieldname }}: {{ $property.Ref | cleanRef }},
    {{- end }}
    {{- end }}
}

impl ToRestString for {{ $classname }} {
    fn to_string(&self) -> String {
	let mut output = String::new();

        {{ $isPreviousField := false }}
        output.push_str("{");
	{{- range $propname, $property := $definition.Properties }}
        {{- $fieldname := $propname | snakeCase }}
        {{ if eq $isPreviousField true }}
        output.push_str(",");
        {{- end }}
        {{- $isPreviousField = true }}

	{{- if eq $property.Type "array" }}
        output.push_str(&format!("\"{{ $propname }}\": [{}],", {
            let vec_string = self.{{ $fieldname }}.iter().map(|x| x.to_string()).collect::<Vec<_>>();
            vec_string.join(", ")
        })); 
	{{- else if eq $property.Type "object" }}
  	output.push_str(&format!("\"{{ $propname }}\": {{"{{"}} {} {{"}}"}}", {
	    let map_string = self
		.{{$fieldname}}
		.iter()
		.map(|(key, value)| format!("\"{}\" = {}", key.to_string(), value.to_string()))
		.collect::<Vec<_>>();
	    map_string.join(", ")
	}));
	{{- else if eq $property.Type "string" }}
	output.push_str(&format!("\"{{ $propname }}\": \"{}\"", self.{{ $fieldname }}));
	{{- else }}
	output.push_str(&format!("\"{{ $propname }}\": {}", self.{{ $fieldname }}.to_string()));
	{{- end }}
	{{- end }}
        output.push_str("}");
	return output;
    }
}
{{- end }}

{{- range $url, $path := .Paths }}
{{- range $method, $operation := $path}}
/// {{ $operation.Summary | stripNewlines }}
pub fn {{ $operation.OperationId | stripOperationPrefix | camelCase | snakeCase	 }}(
{{- if $operation.Security }}
{{- with (index $operation.Security 0) }}
    {{- range $key, $value := . }}
	{{- if eq $key "BasicAuth" }}
    basic_auth_username: &str,
    basic_auth_password: &str,
	{{- else if eq $key "HttpKeyAuth" }}
    bearer_token: &str,
	{{- end }}
    {{- end }}
{{- end }}
{{- else }}
    bearer_token: &str,
{{- end }}
{{- range $parameter := $operation.Parameters }}
{{- $argname := $parameter.Name | snakeCase }}
{{- if eq $parameter.In "path" }}
    {{- if eq $parameter.Type "string" }}
    {{ $argname }}:{{ " " }}{{- if not $parameter.Required }}Option<{{- end }}&str{{- if not $parameter.Required }}>{{- end }}
    {{- else }}
    {{ $argname }}:{{ " " }}{{- if not $parameter.Required }}Option<{{- end }}{{ $parameter.Type }}{{- if not $parameter.Required }}>{{- end }}
    {{- end }}
{{- else if eq $parameter.In "body" }}
    {{- if eq $parameter.Schema.Type "string" }}
    {{ $argname }}:{{ " " }}{{- if not $parameter.Required }}Option<{{- end }}&str{{- if not $parameter.Required }}>{{- end }}
    {{- else }}
    {{ $argname }}:{{ " " }}{{- if not $parameter.Required }}Option<{{- end }}{{ $parameter.Schema.Ref | cleanRef }}{{- if not $parameter.Required }}>{{- end }}
    {{- end }}
{{- else if eq $parameter.Type "array"}}
    {{- if eq $parameter.Items.Type "string" }}
    {{ $argname }}: &[String]
    {{- else }}
    {{ $argname }}: &[{{ $parameter.Items.Type }}]
    {{- end }}
{{- else if eq $parameter.Type "object"}}
    {{- if eq $parameter.AdditionalProperties.Type "string"}}
IDictionary<string, string> {{ $parameter.Name }}
    {{- else if eq $parameter.Items.Type "integer"}}
IDictionary<string, int> {{ $parameter.Name }}
    {{- else if eq $parameter.Items.Type "boolean"}}
IDictionary<string, int> {{ $parameter.Name }}
    {{- else}}
IDictionary<string, {{ $parameter.Items.Type }}> {{ $parameter.Name }}
    {{- end}}
{{- else if eq $parameter.Type "integer" }}
    {{ $argname }}: Option<i32>
{{- else if eq $parameter.Type "boolean" }}
    {{ $argname }}: Option<bool>
{{- else if eq $parameter.Type "string" }}
    {{ $argname }}: Option<&str>
{{- else }}
    {{ $argname }}: Option<{{ $parameter.Type }}>
{{- end }},
{{- end }}
{{- if $operation.Responses.Ok.Schema.Ref }}
) -> RestRequest<{{ $operation.Responses.Ok.Schema.Ref | cleanRef }}> {
{{- else }}
) -> RestRequest<()> {
{{- end }}
    #[allow(unused_mut)]
    let mut urlpath = "{{- $url }}".to_string();

    {{- range $parameter := $operation.Parameters }}
    {{- $argname := $parameter.Name | snakeCase }}
    {{- if eq $parameter.In "path" }}
    urlpath = urlpath.replace("{{- print "{" $parameter.Name "}"}}",{{" "}} {{- $argname }});
    {{- end }}
    {{- end }}

    #[allow(unused_mut)]
    let mut query_params = String::new();

{{- range $parameter := $operation.Parameters }}
    {{- $argname := $parameter.Name | snakeCase }}
{{- if eq $parameter.In "query"}}
    {{- if eq $parameter.Type "integer" }}
if let Some(param) = {{ $argname }} {
    query_params.push_str(&format!("{{- $argname }}={}&", param));
}
    {{- else if eq $parameter.Type "string" }}
if let Some(param) = {{ $argname }} {
    query_params.push_str(&format!("{{- $argname }}={}&", param));
}
    {{- else if eq $parameter.Type "boolean" }}
if let Some(param) = {{ $argname }} {
    query_params.push_str(&format!("{{- $argname }}={:?}&", param));
}
    {{- else if eq $parameter.Type "array" }}
for elem in {{ $argname }}
{
    query_params.push_str(&format!("{{- $argname }}={:?}&", elem));
}
    {{- else }}
{{ $parameter }} // ERROR
    {{- end }}
{{- end }}
{{- end }}

    let authentication = {{- if $operation.Security }}
{{- with (index $operation.Security 0) }}
    {{- range $key, $value := . }}
	{{- if eq $key "BasicAuth" }}
Authentication::Basic {
	username: basic_auth_username.to_owned(),
	password: basic_auth_password.to_owned()
    };
	{{- else if eq $key "HttpKeyAuth" }}
    Authentication::Bearer {
	token: bearer_token.to_owned()
    };
	{{- end }}
    {{- end }}
{{- end }}
{{- else }}
    Authentication::Bearer {
	token: bearer_token.to_owned()
    };
{{- end }}

    {{- $hasBody := false }}
    {{- range $parameter := $operation.Parameters }}
    {{- if eq $parameter.In "body" }}
    {{- $hasBody = true }}
    let body_json = {{ $parameter.Name }}.to_string();
    {{- end }}
    {{- end }}
    {{ if eq $hasBody false }}
    let body_json = String::new();
    {{- end }}

    let method = Method::{{- $method | pascalCase }};

    RestRequest {
       authentication,
       urlpath,
       query_params,
       body: body_json,
       method,
       _marker: std::marker::PhantomData
    }
}

{{- end }}
{{- end }}
`

func convertRefToClassName(input string) (className string) {
	cleanRef := strings.TrimPrefix(input, "#/definitions/")
	className = strings.Title(cleanRef)
	return
}

func snakeCaseToCamelCase(input string) (camelCase string) {
	isToUpper := false
	for k, v := range input {
		if k == 0 {
			camelCase = strings.ToLower(string(input[0]))
		} else {
			if isToUpper {
				camelCase += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					camelCase += string(v)
				}
			}
		}

	}
	return
}

func snakeCaseToPascalCase(input string) (output string) {
	isToUpper := false
	for k, v := range input {
		if k == 0 {
			output = strings.ToUpper(string(input[0]))
		} else {
			if isToUpper {
				output += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					output += string(v)
				}
			}
		}
	}
	return
}

func isSnakeCase(input string) (output bool) {

	output = true

	for _, v := range input {
		vString := string(v)
		if vString != "_" && strings.ToUpper(vString) == vString {
			output = false
		}
	}

	return
}

func camelCaseToSnakeCase(input string) (output string) {
	output = ""

	if isSnakeCase(input) {
		output = input
		return
	}

	for _, v := range input {
		vString := string(v)
		if vString == strings.ToUpper(vString) {
			output += "_" + strings.ToLower(vString)
		} else {
			output += vString
		}
	}

	return
}

func stripNewlines(input string) (output string) {
	output = strings.Replace(input, "\n", " ", -1)
	return
}

func stripOperationPrefix(input string) string {
	return strings.Replace(input, "Nakama_", "", 1)
}

func main() {
	// Argument flags
	var output = flag.String("output", "", "The output for generated code.")
	flag.Parse()

	inputs := flag.Args()
	if len(inputs) < 1 {
		fmt.Printf("No input file found: %s\n\n", inputs)
		fmt.Println("openapi-gen [flags] inputs...")
		flag.PrintDefaults()
		return
	}

	inputFile := inputs[0]
	content, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Unable to read file: %s\n", err)
		return
	}

	var subnamespace (string) = ""

	if len(inputs) > 1 {
		if len(inputs[1]) <= 0 {
			fmt.Println("Empty Sub-Namespace provided.")
			return
		}

		subnamespace = inputs[1]
	}

	var schema struct {
		SubNamespace string
		Paths        map[string]map[string]struct {
			Summary     string
			OperationId string
			Responses   struct {
				Ok struct {
					Schema struct {
						Ref string `json:"$ref"`
					}
				} `json:"200"`
			}
			Parameters []struct {
				Name     string
				In       string
				Required bool
				Type     string   // used with primitives
				Items    struct { // used with type "array"
					Type string
				}
				Schema struct { // used with http body
					Type string
					Ref  string `json:"$ref"`
				}
				Format string // used with type "boolean"
			}
			Security []map[string][]struct {
			}
		}
		Definitions map[string]struct {
			Properties map[string]struct {
				Type  string
				Ref   string   `json:"$ref"` // used with object
				Items struct { // used with type "array"
					Type string
					Ref  string `json:"$ref"`
				}
				AdditionalProperties struct {
					Type string // used with type "map"
				}
				Format      string // used with type "boolean"
				Description string
			}
			Description string
		}
	}

	schema.SubNamespace = subnamespace

	if err := json.Unmarshal(content, &schema); err != nil {
		fmt.Printf("Unable to decode input file %s : %s\n", inputFile, err)
		return
	}

	fmap := template.FuncMap{
		"camelCase":            snakeCaseToCamelCase,
		"cleanRef":             convertRefToClassName,
		"pascalCase":           snakeCaseToPascalCase,
		"stripNewlines":        stripNewlines,
		"title":                strings.Title,
		"uppercase":            strings.ToUpper,
		"snakeCase":            camelCaseToSnakeCase,
		"stripOperationPrefix": stripOperationPrefix,
	}

	tmpl, err := template.New(inputFile).Funcs(fmap).Parse(codeTemplate)
	if err != nil {
		fmt.Printf("Template parse error: %s\n", err)
		return
	}

	if len(*output) < 1 {
		tmpl.Execute(os.Stdout, schema)
		return
	}

	f, err := os.Create(*output)
	if err != nil {
		fmt.Printf("Unable to create file: %s\n", err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	tmpl.Execute(writer, schema)
	writer.Flush()
}