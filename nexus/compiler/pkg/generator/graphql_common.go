// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"k8s.io/utils/strings/slices"

	gqlgenTemplates "github.com/99designs/gqlgen/codegen/templates"
	log "github.com/sirupsen/logrus"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/util"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	//GraphQL standard data types
	Int     string = "Int"
	Float   string = "Float"
	String  string = "String"
	Boolean string = "Boolean"

	// Custom query
	CustomQuerySchema = `Id: ID
	ParentLabels: Map
`
)

type ReturnStatement struct {
	Alias       string
	ReturnType  string
	FieldCount  int
	CRDName     string
	ChainAPI    string
	IsSingleton bool
}

type FieldProperty struct {
	IsResolver              bool
	IsNexusTypeField        bool
	IsNexusOrSingletonField bool
	IsChildOrLink           bool
	IsChildrenOrLinks       bool
	IsMapTypeField          bool
	IsArrayTypeField        bool
	IsStdTypeField          bool
	IsCustomTypeField       bool
	IsPointerTypeField      bool
	IsStringType            bool
	IsArrayStdType          bool
	IsSingleton             bool
	PkgName                 string
	NodeName                string
	// ParentGoTypeName is the gqlgen-normalized Go type name of this field's
	// owning node (= gqlGoTypeName(parentSchemaName)). Templates must use
	// model.<ParentGoTypeName> when emitting references to the parent node's
	// model struct, rather than concatenating PkgName + NodeName.
	ParentGoTypeName string
	FieldName        string
	FieldType        string
	FieldTypePkgPath string
	// GoFieldTypeName is the gqlgen-normalized Go type name for this field's
	// target type. Templates should use model.<GoFieldTypeName> rather than
	// model.<FieldTypePkgPath>; FieldTypePkgPath remains the legacy compiler-
	// internal identifier (used as retMap/linkAPI key) and may diverge from
	// what gqlgen's modelgen actually emits.
	GoFieldTypeName string
	ModelType       string
	SchemaFieldName string
	SchemaTypeName  string
	BaseTypeName    string
	Alias           string
	ReturnType      string
	FieldCount      int
	CRDName         string
	ChainAPI        string
	LinkAPI         string
}

type NodeProperty struct {
	IsParentNode    bool
	HasParent       bool
	IsSingletonNode bool
	IsNexusNode     bool
	BaseImportPath  string
	CrdName         string
	ResolverCount   int
	PkgName         string
	NodeName        string
	SchemaName      string
	// GoTypeName is the Go type name that gqlgen's modelgen emits for this
	// node's schema type (= gqlGoTypeName(SchemaName)). Resolver templates
	// must reference model.<GoTypeName> rather than concatenating PkgName +
	// NodeName, which can diverge from gqlgen's normalization for names
	// containing trailing all-caps tokens (e.g. RTVGPU -> Rtvgpu).
	GoTypeName             string
	Alias                  string
	ReturnType             string
	GroupResourceNameTitle string
	GroupName              string
	Singular               string
	Kind                   string
	ResourceName           string
	ResourceVersion        string
	Singleton              string
	ChildFields            []FieldProperty
	LinkFields             []FieldProperty
	ChildrenFields         []FieldProperty
	LinksFields            []FieldProperty
	ArrayFields            []FieldProperty
	CustomFields           []FieldProperty
	NonStructFields        []FieldProperty
	GraphqlSchemaFields    []FieldProperty
	ResolverFields         map[string][]FieldProperty
	CustomQueries          []nexus.GraphQLQuery
	GraphQlSpec            nexus.GraphQLSpec
}

// Convert go standardType to GraphQL standardType
func convertGraphqlStdType(t string) string {
	// remove pointers
	typeWithoutPointer := strings.ReplaceAll(t, "*", "")
	switch typeWithoutPointer {
	case "string":
		return String
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return Int
	case "bool":
		return Boolean
	case "float32", "float64":
		return Float
	default:
		return ""
	}
}

func jsonMarshalResolver(FieldName, PkgName string) string {
	return fmt.Sprintf("%s, _ := json.Marshal(v%s.Spec.%s)\n%sData := string(%s)\n", FieldName, PkgName, FieldName, FieldName, FieldName)
}

func getPkgName(pkgs parser.Packages, pkgPath string) string {
	importPath, err := strconv.Unquote(pkgPath)
	if err != nil {
		log.Errorf("Failed to parse the package path : %s: %v", pkgPath, err)
	}
	return pkgs[importPath].Name
}

func GetPkg(pkgs parser.Packages, pkgPath string) parser.Package {
	importPath, err := strconv.Unquote(pkgPath)
	if err != nil {
		log.Errorf("Failed to parse the package path : %s: %v", pkgPath, err)
	}
	return pkgs[importPath]
}

// gqlGoTypeName returns the Go type name that gqlgen's modelgen will emit for
// a given GraphQL schema type name. It is the single source of truth for
// translating schema names to Go names; resolver code emitted by the compiler
// must reference model types via this normalization so the symbols line up
// with the structs gqlgen generates in models_gen.go.
func gqlGoTypeName(schemaName string) string {
	return gqlgenTemplates.ToGo(schemaName)
}

func genSchemaResolverName(fn1, fn2 string) (string, string) {
	return fmt.Sprintf("%s_%s", strings.ToLower(fn1), fn2), cases.Title(language.Und, cases.NoLower).String(fn1) + cases.Title(language.Und, cases.NoLower).String(fn2)
}

// genSchemaResolverNames is like genSchemaResolverName but additionally returns
// the gqlgen-normalized Go type name (matching what gqlgen's modelgen will emit
// in models_gen.go). Callers that need to reference the model type from
// generated resolver code should use the third return value; the second return
// value remains the legacy compiler-internal resolver type name used as a key
// in retMap/linkAPI maps.
func genSchemaResolverNames(fn1, fn2 string) (schemaName, resolverTypeName, goTypeName string) {
	schemaName, resolverTypeName = genSchemaResolverName(fn1, fn2)
	goTypeName = gqlGoTypeName(schemaName)
	return
}

// ValidateImportPkg returns the GraphQL schema type name and the legacy
// compiler-internal resolver type name used as a key in retMap/linkAPI maps.
func ValidateImportPkg(pkgName, typeString string, importMap map[string]string, pkgs parser.Packages) (string, string) {
	schemaName, resolverTypeName, _ := ValidateImportPkgWithGoName(pkgName, typeString, importMap, pkgs)
	return schemaName, resolverTypeName
}

// ValidateImportPkgWithGoName extends ValidateImportPkg with a third return
// value: the gqlgen-normalized Go type name for the field's target type.
// Resolver template references like model.<X> must use this third value rather
// than the legacy resolverTypeName so symbols match gqlgen's modelgen output.
func ValidateImportPkgWithGoName(pkgName, typeString string, importMap map[string]string, pkgs parser.Packages) (string, string, string) {
	typeWithoutPointers := strings.ReplaceAll(typeString, "*", "")
	if strings.Contains(typeWithoutPointers, ".") {
		part := strings.Split(typeWithoutPointers, ".")
		if val, ok := importMap[part[0]]; ok {
			resolvedPkgName := getPkgName(pkgs, val)
			repName := strings.ReplaceAll(resolvedPkgName, "-", "")
			return genSchemaResolverNames(repName, part[1])
		}
		for _, v := range importMap {
			resolvedPkgName := getPkgName(pkgs, v)
			if resolvedPkgName == part[0] {
				repName := strings.ReplaceAll(resolvedPkgName, "-", "")
				return genSchemaResolverNames(repName, part[1])
			}
		}
		return genSchemaResolverNames(pkgName, part[1])
	}
	return genSchemaResolverNames(pkgName, typeWithoutPointers)
}

func GetNexusSchemaFieldName(GraphQlSpec nexus.GraphQLSpec) string {
	name := "id"
	value := "ID"
	if GraphQlSpec.IdName != "" {
		name = GraphQlSpec.IdName
	}
	if !GraphQlSpec.IdNullable {
		value = "ID!"
	}
	return fmt.Sprintf("%s: %s", name, value)
}

func GetNodeDetails(pkgName, typeString string, importMap map[string]string, pkgs parser.Packages, gqlSpecMap map[string]nexus.GraphQLSpec) string {
	typeWithoutPointers := strings.ReplaceAll(typeString, "*", "")
	if strings.Contains(typeWithoutPointers, ".") {
		part := strings.Split(typeWithoutPointers, ".")
		if val, ok := importMap[part[0]]; ok {
			p := GetPkg(pkgs, val)
			if val, ok := parser.GetNexusGraphqlSpecAnnotation(p, part[1]); ok {
				gqlspec := gqlSpecMap[fmt.Sprintf("%s.%s", p.Name, val)]
				return GetNexusSchemaFieldName(gqlspec)
			}
		}
		for _, v := range importMap {
			pkgName := getPkgName(pkgs, v)
			if pkgName == part[0] {
				p := GetPkg(pkgs, v)
				if val, ok := parser.GetNexusGraphqlSpecAnnotation(p, part[1]); ok {
					gqlspec := gqlSpecMap[fmt.Sprintf("%s.%s", p.Name, val)]
					return GetNexusSchemaFieldName(gqlspec)
				}
			}
		}
		return "id: ID"
	}
	return "id: ID"
}

func getBaseNodeType(typeString string) string {
	if strings.Contains(typeString, ".") {
		part := strings.Split(typeString, ".")
		return part[1]
	}
	return typeString
}

func constructNexusTypeMap(nodes []*NodeProperty) map[string]string {
	crdNameMap := make(map[string]string)
	for _, n := range nodes {
		if n.IsNexusNode || n.IsSingletonNode {
			crdNameMap[n.CrdName] = n.PkgName + n.NodeName
		}
	}
	return crdNameMap
}

func findTypeAndPkgForField(ptParts []string, importMap map[string]string, pkgs map[string]parser.Package) (string, *parser.Package) {
	structPkg := ptParts[0]
	structType := ptParts[1]

	pkgPath, ok := importMap[structPkg]
	if !ok {
		log.Errorf("Cannot find the package name %s for the type %s", structPkg, structType)
		return "", nil
	}

	importPath, err := strconv.Unquote(pkgPath)
	if err != nil {
		log.Errorf("Failed to parse package %s for the type %s with error %v", pkgPath, structType, err)
		return "", nil
	}

	p, ok := pkgs[importPath]
	if !ok {
		log.Errorf("Cannot find the package details from the path %s for the type %s", importPath, structType)
		return "", nil
	}

	return structType, &p
}

/*
collect and construct type alias field into map recursively before
populating the nexus node and custom struct type
ex. nonStructMap[pkgName] = nodeType  -->  nonStructMap["root"] = "AliasTypeFoo"
*/
func constructAliasType(sortedPackages []parser.Package) map[string]string {
	nonStructMap := make(map[string]string)
	for _, pkg := range sortedPackages {
		for _, node := range pkg.GetNonStructTypes() {
			pkgName := fmt.Sprintf("%s_%s", pkg.Name, parser.GetTypeName(node))
			// NonStruct Map
			nonStructType := types.ExprString(node.Type)
			nonStructMap[pkgName] = nonStructType
		}
	}

	return nonStructMap
}

func setNexusProperties(nodeHelper parser.NodeHelper, node *ast.TypeSpec, nodeProp *NodeProperty) {
	if len(nodeHelper.Parents) > 0 {
		nodeProp.HasParent = true
	}

	if parser.IsSingletonNode(node) {
		nodeProp.IsSingletonNode = true
	}

	if parser.IsNexusNode(node) {
		nodeProp.IsNexusNode = true
	}
}

// isRootOfGraph intended to allow only one root of the graph,
// if we receive multiple node in such behaviour, then we allow the first node and the rest will be skipped with error
// arg: `parents` indicates the node's parents, `rootOfGraph` indicates if the received node is the root of the graph or not.
func isRootOfGraph(parents []string, rootOfGraph bool) bool {
	if len(parents) == 0 && !rootOfGraph {
		return true
	}

	return rootOfGraph
}

func getGraphqlSchemaName(pattern, fieldName, schemaType string, f *ast.Field) string {
	if !strings.HasPrefix(schemaType, "global_") {
		schemaType = cases.Title(language.Und, cases.NoLower).String(schemaType)
	}

	schemaName := fmt.Sprintf(pattern, fieldName, schemaType)
	schemaType = strings.TrimPrefix(schemaType, "global_")
	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_ALIAS_TYPE_ANNOTATION) {
		schemaType = parser.GetFieldAnnotationVal(f, parser.GRAPHQL_ALIAS_TYPE_ANNOTATION)
	}
	if fieldName != "" {
		// use camelCase for fieldName #e.g ServiceGroup --> serviceGroup
		schemaName = fmt.Sprintf(pattern, getAliasFieldValue(fieldName, f), schemaType)
	}

	schemaName = strings.TrimPrefix(schemaName, "global_")

	return schemaName
}

// func getGraphQLArgs(f *ast.Field) string {
// 	argKey := "id"
// 	argVal := "ID"
// 	k := parser.GetGraphqlArgumentKey(f)
// 	if k != "" {
// 		argKey = k
// 	}
// 	v := parser.GetGraphqlArgumentValue(f)
// 	if v != "" {
// 		argVal = v
// 	}
// 	return fmt.Sprintf("%s: %s", argKey, argVal)
// }

// getTsmGraphqlSchemaFieldName process nexus annotation `nexus-graphql-nullable` and `nexus-graphql-tsm-directive`
func getTsmGraphqlSchemaFieldName(sType GraphQLSchemaType, fieldName, schemaType, listArg string, f *ast.Field, pkg parser.Package, nonNexusTypes *parser.NonNexusTypes) string {
	pattern := ""
	nullable := parser.IsNexusGraphqlNullField(f)
	jsonEncoded := parser.IsJsonStringField(f)
	switch sType {
	case Standard, JsonMarshal, Child, Link:
		if nullable {
			pattern = "%s: %s"
		} else {
			pattern = "%s: %s!"
		}
		if jsonEncoded {
			pattern = "%s: %s"
			schemaType = "String"
		}
	case Array:
		if nullable {
			pattern = "%s: [%s]"
		} else {
			pattern = "%s: [%s!]"
		}
		if jsonEncoded {
			pattern = "%s: %s"
			schemaType = "String"
		}
	case NamedChild, NamedLink:
		if nullable {
			pattern = "%s(" + listArg + "): [%s!]"
		} else {
			pattern = "%s(" + listArg + "): [%s]"
		}
		if jsonEncoded {
			pattern = "%s(" + listArg + "): %s"
			schemaType = "String"
		}
	}
	schemaName := getGraphqlSchemaName(pattern, fieldName, schemaType, f)

	if sType == AliasType {
		e := parser.GetFieldAnnotationVal(f, parser.GRAPHQL_ALIAS_TYPE_ANNOTATION)
		if e != "" {
			schemaName = fmt.Sprintf("%s: %s", getAliasFieldValue(fieldName, f), e)
		} else {
			schemaName = fmt.Sprintf("%s: %s", getAliasFieldValue(fieldName, f), "String")
		}
	}

	schemaName = addFieldAnnotations(pkg, f, schemaName, sType, nonNexusTypes)

	return schemaName
}

func addFieldAnnotations(pkg parser.Package, f *ast.Field, schemaName string, sType GraphQLSchemaType, nonNexusTypes *parser.NonNexusTypes) string {
	importMap := pkg.GetImportMap()
	// add jsonencoded annotation
	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_TSM_DIRECTIVE_ANNOTATION) {
		replacer := strings.NewReplacer("nexus-graphql-tsm-directive:", "", "\\", "")
		out := replacer.Replace(parser.GetFieldAnnotationString(f, parser.GRAPHQL_TSM_DIRECTIVE_ANNOTATION))
		schemaName += " " + strings.Trim(out, "\"")
	} else if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_PROTOBUF_NAME) ||
		parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_PROTOBUF_FILE) {
		schemaName = addProtobufAnnotation(f, schemaName)
	} else {
		if sType != Link && sType != Child && sType != NamedChild && sType != NamedLink {
			if val, ok := f.Type.(*ast.SelectorExpr); ok {
				x := types.ExprString(val.X)
				if imp, ok := importMap[x]; ok {
					if strings.HasPrefix(imp, fmt.Sprintf(`"%s`, pkg.ModPath)) {
						schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, val.Sel.Name, schemaName, false, imp)
					} else {
						if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_TYPE_NAME) {
							typeName := parser.GetFieldAnnotationVal(f, parser.GRAPHQL_TYPE_NAME)
							externalType := fmt.Sprintf("%s.%s", x, val.Sel.Name)
							aliasType := fmt.Sprintf("type %s %s", typeName, externalType)
							if nonNexusTypes != nil && !slices.Contains(nonNexusTypes.ExternalTypes, aliasType) {
								nonNexusTypes.ExternalTypes = append(nonNexusTypes.ExternalTypes, aliasType)
							}
							schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, typeName, schemaName, true, imp)
						} else {
							schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, val.Sel.Name, schemaName, true, imp)
						}
					}
				}
			} else if arrayExpr, ok := f.Type.(*ast.ArrayType); ok {
				if val, ok := arrayExpr.Elt.(*ast.SelectorExpr); ok {
					x := types.ExprString(val.X)
					if imp, ok := importMap[x]; ok {
						if strings.HasPrefix(imp, fmt.Sprintf(`"%s`, pkg.ModPath)) {
							schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, val.Sel.Name, schemaName, false, imp)
						} else {
							if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_TYPE_NAME) {
								typeName := parser.GetFieldAnnotationVal(f, parser.GRAPHQL_TYPE_NAME)
								externalType := fmt.Sprintf("%s.%s", x, val.Sel.Name)
								aliasType := fmt.Sprintf("type %s []%s", typeName, externalType)
								if nonNexusTypes != nil && !slices.Contains(nonNexusTypes.ExternalTypes, aliasType) {
									nonNexusTypes.ExternalTypes = append(nonNexusTypes.ExternalTypes, aliasType)
								}
								schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, typeName, schemaName, true, imp)
							} else {
								schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, val.Sel.Name, schemaName, true, imp)
							}
						}
					}
				}
			} else if val, ok := f.Type.(*ast.Ident); ok && convertGraphqlStdType(val.Name) == "" {
				x := ""
				importExpr := ""
				if pkg.Name != "global" {
					x = pkg.Name
					importExpr = fmt.Sprintf(`"%s"`, pkg.FullName)
				}
				schemaName = addJsonencodedAnnotation(f, parser.GRAPHQL_TS_TYPE_ANNOTATION, x, val.Name, schemaName, false, importExpr)
			} else if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_JSONENCODED_ANNOTATION) {
				schemaName += " @jsonencoded"
			}
		}

		schemaName = addRelationAnnotation(sType, f, schemaName)
	}

	return schemaName
}

func addJsonencodedAnnotation(f *ast.Field, annotation parser.FieldAnnotation, x string, name string, schemaName string, external bool, importExpr string) string {
	args := []string{`gofile:"model.go"`, fmt.Sprintf(`name:"%s"`, name)}

	if parser.IsFieldAnnotationPresent(f, annotation) {
		args = append(args, fmt.Sprintf(`file:"%s"`, parser.GetFieldAnnotationVal(f, annotation)))
	}

	if !external && x != "" {
		importStr := strings.Trim(importExpr, "\"")
		namedImport := "nexus_" + importStr[strings.LastIndex(importStr, "/")+1:]
		args = append(args, fmt.Sprintf(`goname:"%s.%s"`, namedImport, name))
	}

	schemaName += fmt.Sprintf(" @jsonencoded(%s)", strings.Join(args, ", "))

	return schemaName
}

func addRelationAnnotation(sType GraphQLSchemaType, f *ast.Field, schemaName string) string {
	var args []string
	if sType == Link || sType == NamedLink {
		args = append(args, `softlink: "true"`)
	}
	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_RELATION_NAME) {
		args = append(args, fmt.Sprintf(`name:"%s"`, parser.GetFieldAnnotationVal(f, parser.GRAPHQL_RELATION_NAME)))
	}

	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_RELATION_PARAMETERS) {
		args = append(args, fmt.Sprintf("parameters:%s", parser.GetFieldAnnotationVal(f, parser.GRAPHQL_RELATION_PARAMETERS)))
	}

	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_RELATION_UUIDKEY) {
		args = append(args, fmt.Sprintf(`uuidkey:"%s"`, parser.GetFieldAnnotationVal(f, parser.GRAPHQL_RELATION_UUIDKEY)))
	}

	if len(args) > 0 {
		schemaName += fmt.Sprintf(" @relation(%s)", strings.Join(args, ", "))
	}

	return schemaName
}

func addProtobufAnnotation(f *ast.Field, schemaName string) string {
	var args []string

	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_PROTOBUF_NAME) {
		args = append(args, fmt.Sprintf(`name:"%s"`, parser.GetFieldAnnotationVal(f, parser.GRAPHQL_PROTOBUF_NAME)))
	}

	if parser.IsFieldAnnotationPresent(f, parser.GRAPHQL_PROTOBUF_FILE) {
		args = append(args, fmt.Sprintf(`file:"%s"`, parser.GetFieldAnnotationVal(f, parser.GRAPHQL_PROTOBUF_FILE)))
	}

	if len(args) > 0 {
		schemaName += fmt.Sprintf(" @protobuf(%s)", strings.Join(args, ", "))
	}

	return schemaName
}

// getAliasFieldValue process nexus annotation  `nexus-alias-value:`
func getAliasFieldValue(fieldName string, f *ast.Field) string {
	e := parser.GetFieldAnnotationVal(f, parser.GRAPHQL_ALIAS_NAME_ANNOTATION)
	if e != "" {
		return e
	}
	return util.GetTag(fieldName)
}
