package v1beta1

import xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"

func (in *Webhook) SetWriteConnectionSecretToReference(r *xpv2.LocalSecretReference) {
	in.Spec.WriteConnectionSecretToReference = r
}
