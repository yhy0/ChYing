// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

import (
	"encoding/xml"
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"log"
	"strings"
	"unicode"
)

const maxRecursion uint8 = 20

// GoWSDL defines the struct for WSDL generator.
type GoWSDL struct {
	loc                   string
	RawWSDL               []byte
	WSDL                  *WSDL
	resolvedXSDExternals  map[string]bool
	currentRecursionLevel uint8
}

func downloadFile(url string) ([]byte, error) {
	response, err := httpx.Get(url, "WSDL")
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Received response code %d", response.StatusCode)
	}

	return []byte(response.Body), nil
}

// NewGoWSDL initializes WSDL generator.
func NewGoWSDL(u string) (*GoWSDL, error) {
	g := &GoWSDL{
		loc: u,
	}

	err := g.unmarshal()
	if err != nil {
		return nil, err
	}

	// Process WSDL nodes
	for _, schema := range g.WSDL.Types.Schemas {
		newTraverser(schema, g.WSDL.Types.Schemas).traverse()
	}

	return g, nil
}

func (g *GoWSDL) fetchFile(url string) (data []byte, err error) {
	data, err = downloadFile(url)

	return
}

func (g *GoWSDL) unmarshal() error {
	data, err := g.fetchFile(g.loc)
	if err != nil {
		return err
	}

	g.WSDL = new(WSDL)
	err = xml.Unmarshal(data, g.WSDL)
	if err != nil {
		return err
	}
	g.RawWSDL = data

	for _, schema := range g.WSDL.Types.Schemas {
		err = g.resolveXSDExternals(schema, g.loc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GoWSDL) resolveXSDExternals(schema *XSDSchema, loc string) error {
	download := func(location string, ref string) error {
		if g.resolvedXSDExternals[location] {
			return nil
		}
		if g.resolvedXSDExternals == nil {
			g.resolvedXSDExternals = make(map[string]bool, maxRecursion)
		}
		g.resolvedXSDExternals[location] = true

		var data []byte
		var err error
		if data, err = g.fetchFile(location); err != nil {
			return err
		}

		newschema := new(XSDSchema)

		err = xml.Unmarshal(data, newschema)
		if err != nil {
			return err
		}

		if (len(newschema.Includes) > 0 || len(newschema.Imports) > 0) &&
			maxRecursion > g.currentRecursionLevel {
			g.currentRecursionLevel++

			err = g.resolveXSDExternals(newschema, location)
			if err != nil {
				return err
			}
		}

		g.WSDL.Types.Schemas = append(g.WSDL.Types.Schemas, newschema)

		return nil
	}

	for _, impts := range schema.Imports {
		// Download the file only if we have a hint in the form of schemaLocation.
		if impts.SchemaLocation == "" {
			log.Printf("[WARN] Don't know where to find XSD for %s", impts.Namespace)
			continue
		}

		if e := download(loc, impts.SchemaLocation); e != nil {
			return e
		}
	}

	for _, incl := range schema.Includes {
		if e := download(loc, incl.SchemaLocation); e != nil {
			return e
		}
	}

	return nil
}

var reservedWords = map[string]string{
	"break":       "break_",
	"default":     "default_",
	"func":        "func_",
	"interface":   "interface_",
	"select":      "select_",
	"case":        "case_",
	"defer":       "defer_",
	"go":          "go_",
	"map":         "map_",
	"struct":      "struct_",
	"chan":        "chan_",
	"else":        "else_",
	"goto":        "goto_",
	"package":     "package_",
	"switch":      "switch_",
	"const":       "const_",
	"fallthrough": "fallthrough_",
	"if":          "if_",
	"range":       "range_",
	"type":        "type_",
	"continue":    "continue_",
	"for":         "for_",
	"import":      "import_",
	"return":      "return_",
	"var":         "var_",
}

var reservedWordsInAttr = map[string]string{
	"break":       "break_",
	"default":     "default_",
	"func":        "func_",
	"interface":   "interface_",
	"select":      "select_",
	"case":        "case_",
	"defer":       "defer_",
	"go":          "go_",
	"map":         "map_",
	"struct":      "struct_",
	"chan":        "chan_",
	"else":        "else_",
	"goto":        "goto_",
	"package":     "package_",
	"switch":      "switch_",
	"const":       "const_",
	"fallthrough": "fallthrough_",
	"if":          "if_",
	"range":       "range_",
	"type":        "type_",
	"continue":    "continue_",
	"for":         "for_",
	"import":      "import_",
	"return":      "return_",
	"var":         "var_",
	"string":      "astring",
}

var specialCharacterMapping = map[string]string{
	"+": "Plus",
	"@": "At",
}

// Replaces Go reserved keywords to avoid compilation issues
func replaceReservedWords(identifier string) string {
	value := reservedWords[identifier]
	if value != "" {
		return value
	}
	return normalize(identifier)
}

// Replaces Go reserved keywords to avoid compilation issues
func replaceAttrReservedWords(identifier string) string {
	value := reservedWordsInAttr[identifier]
	if value != "" {
		return value
	}
	return normalize(identifier)
}

// Normalizes value to be used as a valid Go identifier, avoiding compilation issues
func normalize(value string) string {
	for k, v := range specialCharacterMapping {
		value = strings.ReplaceAll(value, k, v)
	}

	mapping := func(r rune) rune {
		if r == '.' || r == '-' {
			return '_'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}
		return -1
	}

	return strings.Map(mapping, value)
}

func goString(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

var xsd2GoTypes = map[string]string{
	"string":             "string",
	"token":              "string",
	"float":              "float32",
	"double":             "float64",
	"decimal":            "float64",
	"integer":            "int32",
	"int":                "int32",
	"short":              "int16",
	"byte":               "int8",
	"long":               "int64",
	"boolean":            "bool",
	"datetime":           "soap.XSDDateTime",
	"date":               "soap.XSDDate",
	"time":               "soap.XSDTime",
	"base64binary":       "[]byte",
	"hexbinary":          "[]byte",
	"unsignedint":        "uint32",
	"nonnegativeinteger": "uint32",
	"unsignedshort":      "uint16",
	"unsignedbyte":       "byte",
	"unsignedlong":       "uint64",
	"anytype":            "AnyType",
	"ncname":             "NCName",
	"anyuri":             "AnyURI",
}

func removeNS(xsdType string) string {
	// Handles name space, ie. xsd:string, xs:string
	r := strings.Split(xsdType, ":")

	if len(r) == 2 {
		return r[1]
	}

	return r[0]
}

func toGoType(xsdType string, nillable bool) string {
	// Handles name space, ie. xsd:string, xs:string
	r := strings.Split(xsdType, ":")

	t := r[0]

	if len(r) == 2 {
		t = r[1]
	}

	value := xsd2GoTypes[strings.ToLower(t)]

	if value != "" {
		if nillable {
			value = "*" + value
		}
		return value
	}

	return "*" + replaceReservedWords(makePublic(t))
}

// Given a message, finds its type.
//
// I'm not very proud of this function but
// it works for now and performance doesn't
// seem critical at this point
func (g *GoWSDL) FindType(message string) string {
	message = stripns(message)

	for _, msg := range g.WSDL.Messages {
		if msg.Name != message {
			continue
		}

		// Assumes document/literal wrapped WS-I
		if len(msg.Parts) == 0 {
			// Message does not have parts. This could be a Port
			// with HTTP binding or SOAP 1.2 binding, which are not currently
			// supported.
			log.Printf("[WARN] %s message doesn't have any parts, ignoring message...", msg.Name)
			continue
		}

		part := msg.Parts[0]
		if part.Type != "" {
			return stripns(part.Type)
		}

		elRef := stripns(part.Element)

		for _, schema := range g.WSDL.Types.Schemas {
			for _, el := range schema.Elements {
				if strings.EqualFold(elRef, el.Name) {
					if el.Type != "" {
						return stripns(el.Type)
					}
					return el.Name
				}
			}
		}
	}
	return ""
}

// Given a type, check if there's an Element with that type, and return its name.
func (g *GoWSDL) findNameByType(name string) string {
	return newTraverser(nil, g.WSDL.Types.Schemas).findNameByType(name)
}

// TODO(c4milo): Add support for namespaces instead of striping them out
// TODO(c4milo): improve runtime complexity if performance turns out to be an issue.
func (g *GoWSDL) findSOAPAction(operation, portType string) string {
	for _, binding := range g.WSDL.Binding {
		if strings.ToUpper(stripns(binding.Type)) != strings.ToUpper(portType) {
			continue
		}

		for _, soapOp := range binding.Operations {
			if soapOp.Name == operation {
				return soapOp.SOAPOperation.SOAPAction
			}
		}
	}
	return ""
}

func (g *GoWSDL) findServiceAddress(name string) string {
	for _, service := range g.WSDL.Service {
		for _, port := range service.Ports {
			if port.Name == name {
				return port.SOAPAddress.Location
			}
		}
	}
	return ""
}

// TODO(c4milo): Add namespace support instead of stripping it
func stripns(xsdType string) string {
	r := strings.Split(xsdType, ":")
	t := r[0]

	if len(r) == 2 {
		t = r[1]
	}

	return t
}

func makePublic(identifier string) string {
	if isBasicType(identifier) {
		return identifier
	}
	if identifier == "" {
		return "EmptyString"
	}
	field := []rune(identifier)
	if len(field) == 0 {
		return identifier
	}

	field[0] = unicode.ToUpper(field[0])
	return string(field)
}

var basicTypes = map[string]string{
	"string":      "string",
	"float32":     "float32",
	"float64":     "float64",
	"int":         "int",
	"int8":        "int8",
	"int16":       "int16",
	"int32":       "int32",
	"int64":       "int64",
	"bool":        "bool",
	"time.Time":   "time.Time",
	"[]byte":      "[]byte",
	"byte":        "byte",
	"uint16":      "uint16",
	"uint32":      "uint32",
	"uinit64":     "uint64",
	"interface{}": "interface{}",
}

func isBasicType(identifier string) bool {
	if _, exists := basicTypes[identifier]; exists {
		return true
	}
	return false
}
