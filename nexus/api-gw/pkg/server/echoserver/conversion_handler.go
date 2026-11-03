// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ConvertHandler implements the Kubernetes CRD conversion webhook contract
// (https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)
// for Nexus multi-version Kinds (see orch-utils/nexus/compiler
// pkg/parser/node_version.go). CRDs for Kinds with more than one version
// point their spec.conversion.webhook.clientConfig at this endpoint (see
// RenderCRDBaseTemplate / crd_base.yaml.tmpl in the compiler).
//
// Conversion here is a structural no-op by design: Nexus requires every
// version of a Kind to be schema-compatible with the storage version (new
// versions only ever add fields on top of the storage version's shape - see
// the compiler's node_version.go validation), so there is never a field
// rename or semantic transform to apply. Every object is passed through
// unchanged except for `apiVersion`, which is rewritten to whatever version
// the apiserver asked for:
//   - Reading an older stored object at a newer version: fields that don't
//     exist yet on the stored object are simply absent (treated as unset).
//   - Writing a newer-version object down to the storage version: fields the
//     storage version's schema doesn't recognize are pruned by the
//     apiserver itself on write, same as any other unknown-field handling.
//
// Because this handler is generic (nexus-api-gw is datamodel-agnostic and
// only ever sees unstructured JSON - it does not import any specific
// datamodel's generated Go types), it cannot apply custom per-Kind
// conversion semantics (e.g. deriving one field's default value from
// another's). Kinds that need that still have to do it client-side for now;
// teaching this handler to invoke per-Kind, datamodel-specific conversion
// logic is tracked as a follow-up.
func ConvertHandler(c echo.Context) error {
	var review apiextensionsv1.ConversionReview
	if err := json.NewDecoder(c.Request().Body).Decode(&review); err != nil {
		return c.JSON(http.StatusBadRequest, conversionErrorResponse("", fmt.Sprintf("decoding ConversionReview: %v", err)))
	}
	if review.Request == nil {
		return c.JSON(http.StatusBadRequest, conversionErrorResponse("", "ConversionReview.Request is nil"))
	}
	req := review.Request

	desiredGroup := apiVersionGroup(req.DesiredAPIVersion)
	converted := make([]runtime.RawExtension, 0, len(req.Objects))

	for _, raw := range req.Objects {
		var obj map[string]interface{}
		if err := json.Unmarshal(raw.Raw, &obj); err != nil {
			return c.JSON(http.StatusOK, conversionErrorResponse(req.UID, fmt.Sprintf("decoding object: %v", err)))
		}

		srcAPIVersion, _ := obj["apiVersion"].(string)
		if apiVersionGroup(srcAPIVersion) != desiredGroup {
			return c.JSON(http.StatusOK, conversionErrorResponse(req.UID,
				fmt.Sprintf("cannot convert across groups (%s -> %s)", srcAPIVersion, req.DesiredAPIVersion)))
		}

		obj["apiVersion"] = req.DesiredAPIVersion
		convertedRaw, err := json.Marshal(obj)
		if err != nil {
			return c.JSON(http.StatusOK, conversionErrorResponse(req.UID, fmt.Sprintf("encoding converted object: %v", err)))
		}
		converted = append(converted, runtime.RawExtension{Raw: convertedRaw})
	}

	return c.JSON(http.StatusOK, apiextensionsv1.ConversionReview{
		TypeMeta: review.TypeMeta,
		Response: &apiextensionsv1.ConversionResponse{
			UID:              req.UID,
			ConvertedObjects: converted,
			Result:           metav1.Status{Status: metav1.StatusSuccess},
		},
	})
}

func apiVersionGroup(apiVersion string) string {
	idx := strings.LastIndex(apiVersion, "/")
	if idx < 0 {
		return apiVersion
	}
	return apiVersion[:idx]
}

func conversionErrorResponse(uid types.UID, msg string) apiextensionsv1.ConversionReview {
	log.Warn().Msgf("Conversion webhook error: %s", msg)
	return apiextensionsv1.ConversionReview{
		Response: &apiextensionsv1.ConversionResponse{
			UID: uid,
			Result: metav1.Status{
				Status:  metav1.StatusFailure,
				Message: msg,
			},
		},
	}
}
