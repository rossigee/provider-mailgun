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

// Package apis contains Kubernetes API for the Mailgun provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	bouncev1beta1 "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domainv1beta1 "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglistv1beta1 "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routev1beta1 "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialv1beta1 "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatev1beta1 "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	webhookv1beta1 "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
	v1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1beta1.AddToScheme,
		// v1beta1 namespaced versions (Crossplane v2 only)
		bouncev1beta1.AddToScheme,
		domainv1beta1.AddToScheme,
		mailinglistv1beta1.AddToScheme,
		routev1beta1.AddToScheme,
		smtpcredentialv1beta1.AddToScheme,
		templatev1beta1.AddToScheme,
		webhookv1beta1.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
