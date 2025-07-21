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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-mailgun/internal/controller/bounce"
	"github.com/crossplane-contrib/provider-mailgun/internal/controller/domain"
	"github.com/crossplane-contrib/provider-mailgun/internal/controller/mailinglist"
	"github.com/crossplane-contrib/provider-mailgun/internal/controller/route"
	"github.com/crossplane-contrib/provider-mailgun/internal/controller/smtpcredential"
	"github.com/crossplane-contrib/provider-mailgun/internal/controller/template"
	"github.com/crossplane-contrib/provider-mailgun/internal/controller/webhook"
)

// Setup sets up all Mailgun controllers
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		// bounce controllers
		bounce.Setup,
		// domain controllers
		domain.Setup,
		// mailinglist controllers
		mailinglist.Setup,
		// route controllers
		route.Setup,
		// smtpcredential controllers
		smtpcredential.Setup,
		// template controllers
		template.Setup,
		// webhook controllers
		webhook.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// Setup_ProviderConfig sets up the ProviderConfig controller
// Note: ProviderConfig controller setup is handled by crossplane-runtime
