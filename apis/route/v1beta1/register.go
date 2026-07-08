/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RouteAction type metadata.
var (
	RouteActionKind             = reflect.TypeOf(RouteAction{}).Name()
	RouteActionGroupKind        = schema.GroupKind{Group: Group, Kind: RouteActionKind}
	RouteActionKindAPIVersion   = RouteActionKind + "." + SchemeGroupVersion.String()
	RouteActionGroupVersionKind = SchemeGroupVersion.WithKind(RouteActionKind)
)

// Route type metadata.
var (
	RouteKind             = reflect.TypeOf(Route{}).Name()
	RouteGroupKind        = schema.GroupKind{Group: Group, Kind: RouteKind}
	RouteKindAPIVersion   = RouteKind + "." + SchemeGroupVersion.String()
	RouteGroupVersionKind = SchemeGroupVersion.WithKind(RouteKind)
)
