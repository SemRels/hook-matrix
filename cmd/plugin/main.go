// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"context"
	"log"
	"os"

	grpcserver "github.com/SemRels/hook-matrix/internal/grpc"
	semrelplugin "github.com/SemRels/hook-matrix/internal/plugin"
)

func main() {
	provider := semrelplugin.NewMatrixHookFromEnv()
	server := grpcserver.NewProviderServer(provider)

	if _, err := server.Health(context.Background()); err != nil {
		log.Printf("plugin health check failed: %v", err)
		os.Exit(1)
	}

	log.Printf("%s plugin is ready", provider.Name())
}
