// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen, DO NOT EDIT.

package followschema

import (
	"context"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vmware-tanzu/graph-framework-for-microservices/gqlgen/graphql"
)

// region    ************************** generated!.gotpl **************************

// endregion ************************** generated!.gotpl **************************

// region    ***************************** args.gotpl *****************************

// endregion ***************************** args.gotpl *****************************

// region    ************************** directives.gotpl **************************

// endregion ************************** directives.gotpl **************************

// region    **************************** field.gotpl *****************************

// endregion **************************** field.gotpl *****************************

// region    **************************** input.gotpl *****************************

// endregion **************************** input.gotpl *****************************

// region    ************************** interface.gotpl ***************************

// endregion ************************** interface.gotpl ***************************

// region    **************************** object.gotpl ****************************

// endregion **************************** object.gotpl ****************************

// region    ***************************** type.gotpl *****************************

func (ec *executionContext) unmarshalNFallbackToStringEncoding2githubᚗcomᚋ99designsᚋgqlgenᚋcodegenᚋtestserverᚋfollowschemaᚐFallbackToStringEncoding(ctx context.Context, v interface{}) (FallbackToStringEncoding, error) {
	tmp, err := graphql.UnmarshalString(v)
	res := FallbackToStringEncoding(tmp)
	return res, graphql.ErrorOnPath(ctx, err)
}

func (ec *executionContext) marshalNFallbackToStringEncoding2githubᚗcomᚋ99designsᚋgqlgenᚋcodegenᚋtestserverᚋfollowschemaᚐFallbackToStringEncoding(ctx context.Context, sel ast.SelectionSet, v FallbackToStringEncoding) graphql.Marshaler {
	res := graphql.MarshalString(string(v))
	if res == graphql.Null {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
	}
	return res
}

// endregion ***************************** type.gotpl *****************************
