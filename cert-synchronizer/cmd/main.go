// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"

	"r53restapi.com/pkg/apiserver"
	"r53restapi.com/pkg/buildflags"
	"r53restapi.com/pkg/log"
)

const (
	LOG_DIR = "/var/log/oep/aws-management/"
)

var (
	httpPort    = flag.String("port", "8080", "Sets the port to serve on.")
	logLevel    = flag.String("loglevel", "info", "Sets logging level.")
	logFile     = flag.String("logfile", "", "Sets logfile name.")
	logToStdout = flag.Bool("logstdout", true, "Logs to stdout only.")
)

var Commit string

const OEP_VER = "24.08"

func main() {
	flag.Parse()

	logConfig := log.Config{
		Directory:  LOG_DIR,
		File:       *logFile,
		Level:      *logLevel,
		StdOutOnly: *logToStdout,
	}

	if buildflags.DEBUG {
		log.Debugf("***********************************************************************")
		log.Debugf("* Running with debug functionality enabled *")
		log.Debugf("*        Environment Variable Names        *")
		log.Debugf("*          AWS_ACCESS_KEY_ID               *")
		log.Debugf("*         AWS_SECRET_ACCESS_KEY            *")
		log.Debugf("*              AWS_REGION                  *")
		log.Debugf("*               AWS_VPC                    *")
		log.Debugf("*            AWS_R53_DOMAIN                *")
		log.Debugf("*         ACM_CERTIFICATE_NAME             *")
		log.Debugf("*       K8S_CERTIFICATE_NAMESPACE          *")
		log.Debugf("*       DISABLE_CERT_MATCH_CHECKS          *")
		log.Debugf("*          CERTIFICATE_FILE                *")
		log.Debugf("*          PRIVATE_KEY_FILE                *")
		log.Debugf("*         CA_CERTIFICATE_FILE              *")
		log.Debugf("***********************************************************************")
		logConfig.Level = "debug"
	}

	if err := log.Init(logConfig); err != nil {
		log.Errorf("Error initialising logging: %v", err)
		panic("Error initialising logging")
	}

	log.Infof("Starting AWS Management Web endpoint for ACM and Route53")

	ver_info := "Version " + OEP_VER + "_" + Commit

	log.Infof(ver_info)

	serverConf := apiserver.ServerConfig{
		VersionInfo: ver_info,
		HttpPort:    *httpPort,
	}

	if err := apiserver.Init(serverConf); err != nil {
		log.Errorf("Error initialising HTTP server: %v", err)
		panic("Error initialising HTTP server")
	}

	if err := apiserver.Run(); err != nil {
		log.Errorf("Error running HTTP server: %v", err)
		panic("Error running HTTP server")
	}
}
