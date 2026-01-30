package promoter

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// T084: Ent schema file generator implementation
// Convert JSON schema to ent Go code

// SchemaDefinition represents a JSON schema for code generation
type SchemaDefinition struct {
	Type       string                        `json:"type"`
	Properties map[string]PropertyDefinition `json:"properties"`
}

// PropertyDefinition defines a property in the schema
type PropertyDefinition struct {
	Type            string           `json:"type"`
	Required        bool             `json:"required"`
	ValidationRules []ValidationRule `json:"validation_rules,omitempty"`
}

// ValidationRule represents a validation constraint
type ValidationRule struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// FieldDefinition represents a field in the ent schema
type FieldDefinition struct {
	Name            string
	Type            string
	Required        bool
	ValidationRules []ValidationRule
}

// MapFieldType converts JSON schema type to ent field type
func MapFieldType(jsonType string) string {
	switch jsonType {
	case "string":
		return "String"
	case "integer":
		return "Int"
	case "number":
		return "Float"
	case "boolean":
		return "Bool"
	default:
		return "String"
	}
}

// ConvertValidationRules converts validation rules to ent validator calls
func ConvertValidationRules(rules []ValidationRule) []string {
	validators := []string{}

	for _, rule := range rules {
		switch rule.Type {
		case "format":
			if rule.Value == "email" {
				validators = append(validators, "Match(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`))")
			}
		case "minLength":
			if val, ok := rule.Value.(float64); ok {
				validators = append(validators, fmt.Sprintf("MinLen(%d)", int(val)))
			} else if val, ok := rule.Value.(int); ok {
				validators = append(validators, fmt.Sprintf("MinLen(%d)", val))
			}
		case "maxLength":
			if val, ok := rule.Value.(float64); ok {
				validators = append(validators, fmt.Sprintf("MaxLen(%d)", int(val)))
			} else if val, ok := rule.Value.(int); ok {
				validators = append(validators, fmt.Sprintf("MaxLen(%d)", val))
			}
		case "minimum":
			if val, ok := rule.Value.(float64); ok {
				validators = append(validators, fmt.Sprintf("Min(%d)", int(val)))
			} else if val, ok := rule.Value.(int); ok {
				validators = append(validators, fmt.Sprintf("Min(%d)", val))
			}
		}
	}

	return validators
}

// GenerateFieldDefinitions creates field definitions for the ent schema
func GenerateFieldDefinitions(schema SchemaDefinition) []FieldDefinition {
	fields := []FieldDefinition{}

	for propName, propDef := range schema.Properties {
		field := FieldDefinition{
			Name:            propName,
			Type:            MapFieldType(propDef.Type),
			Required:        propDef.Required,
			ValidationRules: propDef.ValidationRules,
		}
		fields = append(fields, field)
	}

	return fields
}

// entSchemaTemplate is the Go template for generating ent schema files
const entSchemaTemplate = `package schema

import (
{{- if .NeedsRegexp }}
	"regexp"

{{- end }}
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// {{ .TypeName }} holds the schema definition for the {{ .TypeName }} entity.
type {{ .TypeName }} struct {
	ent.Schema
}

// Fields of the {{ .TypeName }}.
func ({{ .TypeName }}) Fields() []ent.Field {
	return []ent.Field{
{{- range .Fields }}
		field.{{ .Type }}("{{ .Name }}"){{- if not .Required }}.
			Optional(){{- else }}.
			NotEmpty(){{- end }}{{- range .Validators }}.
			{{ . }}{{- end }},
{{- end }}
	}
}

// Edges of the {{ .TypeName }}.
func ({{ .TypeName }}) Edges() []ent.Edge {
	return nil
}
`

// TemplateData holds data for the ent schema template
type TemplateData struct {
	TypeName    string
	NeedsRegexp bool
	Fields      []struct {
		Name       string
		Type       string
		Required   bool
		Validators []string
	}
}

// GenerateEntSchemaFile generates an ent schema file from a schema definition
func GenerateEntSchemaFile(schema SchemaDefinition, outputDir string) error {
	// Generate field definitions
	fieldDefs := GenerateFieldDefinitions(schema)

	// Prepare template data
	data := TemplateData{
		TypeName:    strings.Title(schema.Type),
		NeedsRegexp: false,
		Fields: make([]struct {
			Name       string
			Type       string
			Required   bool
			Validators []string
		}, 0, len(fieldDefs)),
	}

	for _, field := range fieldDefs {
		validators := ConvertValidationRules(field.ValidationRules)

		// Check if any validator uses regexp
		for _, v := range validators {
			if strings.Contains(v, "regexp.MustCompile") {
				data.NeedsRegexp = true
				break
			}
		}

		templateField := struct {
			Name       string
			Type       string
			Required   bool
			Validators []string
		}{
			Name:       field.Name,
			Type:       field.Type,
			Required:   field.Required,
			Validators: validators,
		}
		data.Fields = append(data.Fields, templateField)
	}

	// Parse template
	tmpl, err := template.New("entschema").Parse(entSchemaTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write to file
	filename := filepath.Join(outputDir, strings.ToLower(schema.Type)+".go")
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
