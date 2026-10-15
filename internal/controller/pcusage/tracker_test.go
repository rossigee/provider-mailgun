package pcusage

import (
	"testing"

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	smtpcredentialv1beta1 "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	v1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
)

func TestNewTracker(t *testing.T) {
	scheme := runtime.NewScheme()
	tracker := NewTracker(nil, scheme)

	assert.NotNil(t, tracker)
	assert.Nil(t, tracker.kube)
	assert.NotNil(t, tracker.scheme)
}

func TestGetProviderConfigReference(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, smtpcredentialv1beta1.SchemeBuilder.AddToScheme(scheme))
	tracker := NewTracker(nil, scheme)

	t.Run("returns reference when type supports it", func(t *testing.T) {
		cr := &smtpcredentialv1beta1.SMTPCredential{}
		cr.Spec.ProviderConfigReference = &xpv1.ProviderConfigReference{Name: "my-pc"}
		ref := tracker.getProviderConfigReference(cr)
		require.NotNil(t, ref)
		assert.Equal(t, "my-pc", ref.Name)
	})

	t.Run("returns nil when reference is nil", func(t *testing.T) {
		cr := &smtpcredentialv1beta1.SMTPCredential{}
		cr.Spec.ProviderConfigReference = nil
		ref := tracker.getProviderConfigReference(cr)
		assert.Nil(t, ref)
	})

	t.Run("returns nil for type without provider config reference method", func(t *testing.T) {
		cr := &smtpcredentialv1beta1.SMTPCredential{}
		ref := tracker.getProviderConfigReference(cr)
		assert.Nil(t, ref)
	})
}
