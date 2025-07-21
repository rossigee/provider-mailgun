//go:build tools
// +build tools

package tools

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "github.com/crossplane/crossplane-tools/cmd/angryjet"
)
