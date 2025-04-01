// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

func NewProxyAWSHandler(svc secretsmanageriface.SecretsManagerAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		secretName := r.URL.Query().Get("name")
		if secretName == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "query param name empty")
			return
		}
		log.Println("handling request for secret:", secretName)
		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretName),
		}

		result, err := svc.GetSecretValue(input)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			return
		}

		if result.SecretString != nil {
			fmt.Fprintln(w, *result.SecretString)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "secret is binary")
		}
	}
}
