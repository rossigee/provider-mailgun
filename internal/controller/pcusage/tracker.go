package pcusage

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-mailgun/apis/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errTrackPCUsage      = "cannot track ProviderConfig usage"
	errGetProviderConfig = "cannot get ProviderConfig"
	errCreatePCUsage     = "cannot create ProviderConfigUsage"
	errGetPCUsage        = "cannot get ProviderConfigUsage"
)

type Tracker struct {
	kube   client.Client
	scheme *runtime.Scheme
}

func NewTracker(kube client.Client, scheme *runtime.Scheme) *Tracker {
	return &Tracker{
		kube:   kube,
		scheme: scheme,
	}
}

func (t *Tracker) Track(ctx context.Context, mg resource.ModernManaged) error {
	pcRef := mg.GetProviderConfigReference()
	if pcRef == nil {
		return nil
	}

	pc := &v1beta1.ProviderConfig{}
	if err := t.kube.Get(ctx, types.NamespacedName{Name: pcRef.Name}, pc); err != nil {
		return errors.Wrap(err, errGetProviderConfig)
	}

	pcu := &v1beta1.ProviderConfigUsage{}
	pcUsageKey := types.NamespacedName{
		Name: fmt.Sprintf("%s-%s", mg.GetName(), mg.GetUID()),
	}
	if err := t.kube.Get(ctx, pcUsageKey, pcu); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, errGetPCUsage)
		}

		pcu = &v1beta1.ProviderConfigUsage{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1beta1.SchemeGroupVersion.String(),
				Kind:       v1beta1.ProviderConfigUsageKind,  // from register.go
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      pcUsageKey.Name,
				Namespace: mg.GetNamespace(),
			},
			ProviderConfigUsage: xpv1.ProviderConfigUsage{
				ProviderConfigReference: xpv1.Reference{
					Name: pcRef.Name,
				},
				ResourceReference: xpv1.TypedReference{
					APIVersion: mg.GetObjectKind().GroupVersionKind().GroupVersion().String(),
					Kind:       mg.GetObjectKind().GroupVersionKind().Kind,
					Name:       mg.GetName(),
					UID:        mg.GetUID(),
				},
			},
		}

		if err := t.kube.Create(ctx, pcu); err != nil {
			return errors.Wrap(err, errCreatePCUsage)
		}
	}

	return nil
}
