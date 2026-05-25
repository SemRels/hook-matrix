// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"log"

	plugin "github.com/SemRels/hook-matrix/internal/plugin"
)

func main() {
	notifier := plugin.NewMatrixNotifier(plugin.MatrixConfig{})
	log.Printf("hook-matrix plugin ready: sends Matrix release notifications (%T)", notifier)
}
