// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/open-edge-platform/orch-utils/token-fs/internal"
)

const port = ":8080"

func ParseFlags() (string, string, string, bool) {
	var jwksURL string
	flag.StringVar(&jwksURL, "jwksURL", "", "jwksURL endpoint contains public key for input token validation")
	var fileServerPath string
	flag.StringVar(&fileServerPath, "fileServerPath", "", "Release Service token directory")
	var rolesFile string
	flag.StringVar(&rolesFile, "rolesFile", "",
		"roles file holds a list of roles (one per line) that grant access to read Release Service token")
	var emptyRSToken bool
	flag.BoolVar(&emptyRSToken, "emptyRSToken", false, "if true, it will return empty RS token with HTTP 204 status.")
	flag.Parse()
	return jwksURL, fileServerPath, rolesFile, emptyRSToken
}

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	jwksURL, fileServerPath, rolesFile, emptyRSToken := ParseFlags()

	if jwksURL == "" {
		fmt.Println("Missing required -jwksURL flag")
		exitCode = 1
		return
	}
	if rolesFile == "" {
		fmt.Println("Missing required -rolesFile flag")
		exitCode = 1
		return
	}
	if !emptyRSToken && fileServerPath == "" {
		fmt.Println("Missing required -fileServerPath flag")
		exitCode = 1
		return
	}

	// setup auto refresh of jwks URL
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := jwk.NewCache(ctx)
	if err := c.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute)); err != nil {
		fmt.Printf("failed to register jwks URL %s: %v", jwksURL, err)
		exitCode = 1
		return
	}
	// obtain initial keyset
	keySet, err := c.Get(ctx, jwksURL)
	if err != nil {
		fmt.Printf("failed to fetch keyset from jwks URL %s: %v", jwksURL, err)
		exitCode = 1
		return
	}

	// setup file server for reading the RS token
	fs := http.FileServer(http.Dir(fileServerPath))

	// setup roles that will have access to the file server contents
	content, err := os.ReadFile(rolesFile)
	if err != nil {
		fmt.Printf("error reading roles file %s: %v", rolesFile, err)
		exitCode = 1
		return
	}
	roles := strings.Split(string(content), "\n")
	http.HandleFunc("/", internal.NewFileHandler(keySet, fs, roles, emptyRSToken))
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hihi"))
	})

	log.Println("Listening on", port)
	server := &http.Server{
		Addr:              port,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error serving: %v", err)
			exitCode = 1
			return
		}
	}
}
