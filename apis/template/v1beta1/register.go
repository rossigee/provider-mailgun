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

// TemplateVersion type metadata.
var (
	TemplateVersionKind             = reflect.TypeOf(TemplateVersion{}).Name()
	TemplateVersionGroupKind        = schema.GroupKind{Group: Group, Kind: TemplateVersionKind}
	TemplateVersionKindAPIVersion   = TemplateVersionKind + "." + SchemeGroupVersion.String()
	TemplateVersionGroupVersionKind = SchemeGroupVersion.WithKind(TemplateVersionKind)
)

// Template type metadata.
var (
	TemplateKind             = reflect.TypeOf(Template{}).Name()
	TemplateGroupKind        = schema.GroupKind{Group: Group, Kind: TemplateKind}
	TemplateKindAPIVersion   = TemplateKind + "." + SchemeGroupVersion.String()
	TemplateGroupVersionKind = SchemeGroupVersion.WithKind(TemplateKind)
)
