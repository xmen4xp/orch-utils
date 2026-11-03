// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"go/ast"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/util"
)

func generateNexusClientVars(baseGroupName, crdModulePath string, pkgs parser.Packages, parentsMap map[string]parser.NodeHelper) (clientVars, error) {
	var vars clientVars

	vars.BaseClientsetImport = `"` + crdModulePath + `client/clientset/versioned"`
	vars.FakeBaseCliensetImport = `"` + crdModulePath + `client/clientset/versioned/fake"`
	vars.CommonImport = util.GetInternalImport(crdModulePath, "common")
	vars.HelperImport = util.GetInternalImport(crdModulePath, "helper")

	sortedKeys := make([]string, 0, len(pkgs))
	for k := range pkgs {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	sortedPackages := make([]parser.Package, len(pkgs))
	for i, k := range sortedKeys {
		sortedPackages[i] = pkgs[k]
	}

	for _, pkg := range sortedPackages {
		kindOrder, kindGroups := parser.GroupNodesByKind(pkg)
		if len(kindOrder) == 0 {
			continue
		}

		// Every version declared by any Kind in this package gets its own
		// typed client group (e.g. client.Root() for v1, client.RootV2() for
		// v2), mirroring k8s client-go's per-(group,version) clientsets. The
		// v1 accessor keeps its original unqualified name (e.g. "Root", not
		// "RootV1") so every existing caller of a single-version group's
		// client keeps compiling and working unchanged.
		for _, version := range parser.PackageVersions(pkg) {
			var groupVars ApiGroupsVars

			importPath := util.GetImportPath(pkg.Name, baseGroupName, version)
			baseImportName := util.GetBaseImportName(pkg.Name, baseGroupName, version)
			informerImportName := util.GetInformerImportName(pkg.Name, baseGroupName, version)
			vars.BaseImports += baseImportName + ` "` + crdModulePath + "apis/" + importPath + `"` +
				"\n" // eg baseroothelloworldv1 "helloworld/nexus/generated/apis/root.helloworld.com/v1"
			vars.InformerImports += informerImportName + ` "` + crdModulePath + "client/informers/externalversions/" + importPath + `"` +
				"\n" // eg informerroothelloworldv1 "helloworld/nexus/generated/client/informers/externalversions/root.helloworld.com/v1"
			groupVarName := util.GetGroupVarName(pkg.Name, baseGroupName, version)
			groupTypeName := util.GetGroupTypeName(pkg.Name, baseGroupName, version)
			accessorName := util.GetSimpleGroupTypeName(pkg.Name)
			if version != "v1" {
				accessorName += cases.Title(language.Und, cases.NoLower).String(version) // eg RootV2
			}
			groupVars.SimpleGroupTypeName = accessorName
			groupVars.ClientsetApiGroups = groupVarName + " *" + groupTypeName + "\n" // eg rootHelloworldV1 *RootHelloworldV1

			initClient := "client." + groupVarName + " = new" + groupTypeName +
				"(client)\n" // eg client.rootHelloworldV1 = newRootHelloworldV1(client)
			groupVars.InitApiGroups = initClient

			clientsetMethod := "func (c *Clientset) " + accessorName + "() *" + groupTypeName + " {\n" + "return c." +
				groupVarName + "\n}\n" // eg
			// func (c *Clientset) Root() *RootHelloworldV1 {
			//	return c.rootHelloworldV1
			// }

			groupVars.ClientsetsApiGroupMethods = clientsetMethod
			groupVars.GroupTypeName = groupTypeName

			for _, kind := range kindOrder {
				var kv *parser.KindVersion
				for i := range kindGroups[kind] {
					if kindGroups[kind][i].Version == version {
						kv = &kindGroups[kind][i]
						break
					}
				}
				if kv == nil {
					// This Kind doesn't have a struct for this version - it
					// isn't served at this version.
					continue
				}

				var clientGroupVars apiGroupsClientVars
				clientGroupVars.ApiGroupsVars = groupVars
				clientGroupVars.GroupTypeName = groupTypeName
				clientGroupVars.CrdName = util.GetCrdName(kind, pkg.Name, baseGroupName)
				err := resolveNode(baseImportName, informerImportName, pkg, pkgs, baseGroupName, version, &clientGroupVars, kind, kv.TypeSpec, kv.IsStorageVersion, parentsMap)
				if err != nil {
					return clientVars{}, err
				}
				vars.Nodes = append(vars.Nodes, clientGroupVars)
			}

			vars.ApiGroups = append(vars.ApiGroups, groupVars)
		}
	}
	return vars, nil
}

func resolveNode(baseImportName, informerImportName string, pkg parser.Package, allPkgs parser.Packages, baseGroupName, version string,
	clientGroupVars *apiGroupsClientVars, kind string, node *ast.TypeSpec, isStorageVersion bool, parentsMap map[string]parser.NodeHelper) error {
	pkgName := pkg.Name
	baseNodeName := kind // eg Root
	groupResourceNameTitle := util.GetGroupResourceNameTitle(baseNodeName)
	groupResourceType := util.GetGroupResourceType(baseNodeName, pkgName, baseGroupName, version)

	clientGroupVars.BaseNodeName = baseNodeName
	clientGroupVars.BaseImportName = baseImportName
	clientGroupVars.IsSingleton = parser.IsSingletonNode(node)
	clientGroupVars.GroupBaseImport = baseImportName + "." + baseNodeName
	clientGroupVars.GroupInformerImport = informerImportName
	clientGroupVars.GroupResourceType = groupResourceType
	clientGroupVars.GroupResourceNameTitle = groupResourceNameTitle
	if _, ok := parser.GetNexusDeferredDeleteAnnotation(pkg, parser.GetTypeName(node)); ok {
		clientGroupVars.DeferredDelete = true
	}

	//Get user defined status field if present
	statusField := parser.GetStatusField(node)
	if statusField != nil {
		clientGroupVars.HasStatus = true
		clientGroupVars.StatusType = parser.GetFieldType(statusField)
		statusName, err := parser.GetFieldName(statusField)
		if err != nil {
			log.Fatalf("failed to determine field name: %v", err)
		}
		if statusName == "" {
			log.Fatalf("name of the user defined status field in nexus node can't be empty")
		}
		clientGroupVars.StatusName = statusName
		clientGroupVars.StatusNameFirstLower = util.GetTag(statusName)
	} else {
		clientGroupVars.HasStatus = false
	}

	// TODO support resolution of links which are not nexus nodes https://jira.eng.vmware.com/browse/NPT-112
	//
	// Children/links are graph edges, and the graph (pkg/parser/node_parser.go
	// processNode) only ever walks them from a Kind's storage-version struct -
	// a non-storage version's own child/link field declarations (if any) are
	// not wired into parentsMap, so don't generate navigation methods for them
	// here either; only the storage version's client exposes child/link
	// navigation. Non-storage versions still get Get/Update/Delete on the
	// resource itself via the fields loop below.
	if isStorageVersion {
		children := parser.GetChildFields(node)
		childrenAndLinks := children
		childrenAndLinks = append(childrenAndLinks, parser.GetLinkFields(node)...)

		if len(children) > 0 {
			clientGroupVars.HasChildren = true
		}
		for _, link := range childrenAndLinks {
			linkInfo := getFieldInfo(pkg, allPkgs, link)
			// Children/links always point at the referenced Kind's storage
			// version - that's the only version guaranteed to exist for
			// every Kind, including ones that were never given their own
			// extra versions.
			childVersion := storageVersionForKind(allPkgs, linkInfo.pkgName, linkInfo.fieldType)
			clientVarsLink := apiGroupsClientVarsLink{
				FieldName:              linkInfo.fieldName,
				FieldNameGvk:           util.GetGvkFieldTagName(linkInfo.fieldName),
				Group:                  util.GetGroupName(linkInfo.pkgName, baseGroupName),
				Kind:                   linkInfo.fieldType,
				GroupBaseImport:        util.GetBaseImportName(linkInfo.pkgName, baseGroupName, childVersion) + "." + linkInfo.fieldType,
				GroupResourceNameTitle: util.GetGroupResourceNameTitle(linkInfo.fieldType),
				GroupTypeName:          util.GetGroupTypeName(linkInfo.pkgName, baseGroupName, childVersion),
				SimpleGroupTypeName:    util.GetSimpleGroupTypeName(linkInfo.pkgName),
				BaseNodeName:           linkInfo.fieldType,
				GroupResourceType:      util.GetGroupResourceType(linkInfo.fieldType, linkInfo.pkgName, baseGroupName, childVersion),
				CrdName:                util.GetCrdName(linkInfo.fieldType, linkInfo.pkgName, baseGroupName),
				IsSingleton:            isChildSingleton(pkg, allPkgs, link),
			}
			if parser.IsNamedChildOrLink(link) {
				clientVarsLink.IsNamed = true
			} else {
				clientVarsLink.IsNamed = false
			}
			if parser.IsLinkField(link) {
				clientGroupVars.Links = append(clientGroupVars.Links, clientVarsLink)
			} else {
				clientGroupVars.Children = append(clientGroupVars.Children, clientVarsLink)
			}
			clientGroupVars.LinksAndChildren = append(clientGroupVars.LinksAndChildren, clientVarsLink)
		}
	}

	for _, f := range parser.GetSpecFields(node) {
		fieldInfo := getFieldInfo(pkg, allPkgs, f)
		var vars apiGroupsClientVarsLink
		vars.FieldName = fieldInfo.fieldName

		vars.FieldNameTag = parser.GetFieldNameJsonTag(f)
		if vars.FieldNameTag == "" {
			vars.FieldNameTag = util.GetTag(fieldInfo.fieldName)
		}

		clientGroupVars.Fields = append(clientGroupVars.Fields, vars)
	}

	nodeHelper := parentsMap[clientGroupVars.CrdName]
	clientGroupVars.Group = util.GetGroupFromCrdName(clientGroupVars.CrdName)
	clientGroupVars.Kind = nodeHelper.Name
	if len(nodeHelper.Parents) > 0 {
		parentCrdName := nodeHelper.Parents[len(nodeHelper.Parents)-1]
		parentHelper := parentsMap[parentCrdName]

		parentPkgName := util.GetPackageNameFromCrdName(parentCrdName)
		// The parent is always referenced via its own Kind's storage
		// version - same reasoning as the children/links case above.
		parentVersion := storageVersionForKind(allPkgs, parentPkgName, parentHelper.Name)

		clientGroupVars.Parent.HasParent = true
		clientGroupVars.Parent.IsNamed = parentHelper.Children[clientGroupVars.CrdName].IsNamed
		clientGroupVars.Parent.CrdName = parentCrdName
		clientGroupVars.Parent.GroupTypeName = util.GetGroupTypeName(
			parentPkgName, baseGroupName, parentVersion)
		clientGroupVars.Parent.SimpleGroupTypeName = util.GetSimpleGroupTypeName(parentPkgName)
		clientGroupVars.Parent.GroupResourceNameTitle = util.GetGroupResourceNameTitle(parentHelper.Name)
		clientGroupVars.Parent.GvkFieldName = parentHelper.Children[clientGroupVars.CrdName].FieldNameGvk
		clientGroupVars.Parent.GoGvkFieldName = parentHelper.Children[clientGroupVars.CrdName].GoFieldNameGvk
		clientGroupVars.Parent.BaseNodeName = parentHelper.Name
	}

	return nil
}

// storageVersionForKind returns the version of pkgName's storage-version
// struct for kind. Children, links, and parents are always referenced via
// the other Kind's storage version (the one guaranteed to exist for every
// Kind, versioned or not) - never via whichever version of the *current*
// Kind happens to be getting generated. Defaults to "v1" if pkgName/kind
// can't be resolved, matching today's single-version behavior.
func storageVersionForKind(allPkgs parser.Packages, pkgName, kind string) string {
	for _, p := range allPkgs {
		if p.Name != pkgName {
			continue
		}
		_, groups := parser.GroupNodesByKind(p)
		versions, ok := groups[kind]
		if !ok {
			continue
		}
		if sv := parser.StorageVersion(versions); sv != nil {
			return sv.Version
		}
	}
	return "v1"
}

type fieldInfo struct {
	pkgName   string
	fieldName string
	fieldType string
}

func isChildSingleton(pkg parser.Package, allPkgs parser.Packages, f *ast.Field) bool {
	chType := parser.GetFieldType(f)
	split := strings.Split(chType, ".")
	if len(split) > 1 { // imported node
		fieldPackageName := split[0]
		for _, imp := range pkg.GetImports() { // go through imports to find matching package
			var packageNameToCheck string
			if imp.Name != nil { // named import
				packageNameToCheck = imp.Name.String()
			} else {
				unquotedImport, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					continue
				}
				spl := strings.Split(unquotedImport, "/")
				packageNameToCheck = spl[len(spl)-1]
			}
			if fieldPackageName == packageNameToCheck {
				unquotedImport, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					continue
				}
				for _, p := range allPkgs {
					if unquotedImport == p.FullName {
						for _, node := range p.GetNexusNodes() {
							if parser.GetTypeName(node) == split[1] {
								return parser.IsSingletonNode(node) // we found node definition, check if it's singleton
							}
						}
					}
				}
			}
		}
	} else { // node from same package
		for _, node := range pkg.GetNexusNodes() {
			if parser.GetTypeName(node) == chType {
				return parser.IsSingletonNode(node)
			}
		}
	}
	return false
}

func getFieldInfo(pkg parser.Package, allPkgs parser.Packages, f *ast.Field) fieldInfo {
	var info fieldInfo
	var err error
	info.fieldName, err = parser.GetNodeFieldName(f)
	if err != nil {
		log.Fatalf("Failed to get name of field: %v", err)
	}
	currentPkgName := pkg.Name

	chType := parser.GetFieldType(f)
	s := strings.Split(chType, ".")
	info.fieldType = s[len(s)-1]

	split := strings.Split(chType, ".")
	// overwrite pkg name of links or children from different packages
	if len(split) > 1 {
		info.pkgName = split[0]
		// overwrite pkg name for node which uses named import like 'sg "helloworld.com/service-groups"'
		for _, imp := range pkg.GetImports() {
			if imp.Name.String() == info.pkgName {
				// look for import matching package
				for _, p := range allPkgs {
					unquotedImport, err := strconv.Unquote(imp.Path.Value)
					if err != nil {
						continue
					}
					if unquotedImport == p.FullName {
						info.pkgName = p.Name
						break
					}
				}
			}
		}
		info.pkgName = util.RemoveSpecialChars(info.pkgName)
	} else {
		info.pkgName = util.RemoveSpecialChars(currentPkgName)
	}
	return info
}

type ApiGroupsVars struct {
	InitApiGroups             string
	ClientsetApiGroups        string
	ClientsetsApiGroupMethods string
	SimpleGroupTypeName       string
	GroupTypeName             string
	GroupResourcesInit        string
	GroupResources            string
	GroupResourcesDefs        string
}

type apiGroupsClientVars struct {
	ApiGroupsVars
	BaseNodeName           string
	CrdName                string
	DeferredDelete         bool
	IsSingleton            bool
	HasChildren            bool
	HasStatus              bool
	StatusName             string
	StatusNameFirstLower   string
	StatusType             string
	BaseImportName         string
	GroupResourceType      string
	GroupResourceNameTitle string
	GroupBaseImport        string
	GroupInformerImport    string
	Group                  string
	Kind                   string

	Parent struct {
		IsNamed                bool
		HasParent              bool
		CrdName                string
		GvkFieldName           string
		GoGvkFieldName         string
		GroupTypeName          string
		SimpleGroupTypeName    string
		GroupResourceNameTitle string
		BaseNodeName           string
	}

	Links            []apiGroupsClientVarsLink
	Children         []apiGroupsClientVarsLink
	LinksAndChildren []apiGroupsClientVarsLink
	Fields           []apiGroupsClientVarsLink
}

type apiGroupsClientVarsLink struct {
	FieldName              string
	FieldNameTag           string
	FieldNameGvk           string
	Group                  string
	Kind                   string
	GroupBaseImport        string
	BaseNodeName           string
	IsNamed                bool
	IsSingleton            bool
	GroupTypeName          string
	SimpleGroupTypeName    string
	GroupResourceNameTitle string
	GroupResourceType      string
	CrdName                string
}
