package webhooks

import (
	"context"
	"fmt"
	"net/netip"

	"go4.org/netipx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/telekom/cluster-api-ipam-provider-in-cluster/api/v1alpha2"
	"github.com/telekom/cluster-api-ipam-provider-in-cluster/internal/poolutil"
	"github.com/telekom/cluster-api-ipam-provider-in-cluster/pkg/types"
)

func (webhook *InClusterIPPool) SetupWebhookWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha2.InClusterIPPool{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
	if err != nil {
		return err
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha2.GlobalInClusterIPPool{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipam-cluster-x-k8s-io-v1alpha2-inclusterippool,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=ipam.cluster.x-k8s.io,resources=inclusterippools,versions=v1alpha2,name=validation.inclusterippool.ipam.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-ipam-cluster-x-k8s-io-v1alpha2-inclusterippool,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=ipam.cluster.x-k8s.io,resources=inclusterippools,versions=v1alpha2,name=default.inclusterippool.ipam.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipam-cluster-x-k8s-io-v1alpha2-globalinclusterippool,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=ipam.cluster.x-k8s.io,resources=globalinclusterippools,versions=v1alpha2,name=validation.globalinclusterippool.ipam.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-ipam-cluster-x-k8s-io-v1alpha2-globalinclusterippool,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=ipam.cluster.x-k8s.io,resources=globalinclusterippools,versions=v1alpha2,name=default.globalinclusterippool.ipam.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// InClusterIPPool implements a validating and defaulting webhook for InClusterIPPool and GlobalInClusterIPPool.
type InClusterIPPool struct {
	Client client.Reader
}

var _ webhook.CustomDefaulter = &InClusterIPPool{}
var _ webhook.CustomValidator = &InClusterIPPool{}

// Default satisfies the defaulting webhook interface.
func (webhook *InClusterIPPool) Default(_ context.Context, obj runtime.Object) error {
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *InClusterIPPool) ValidateCreate(_ context.Context, obj runtime.Object) error {
	pool, ok := obj.(types.GenericInClusterPool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a InClusterIPPool or an GlobalInClusterIPPool but got a %T", obj))
	}
	return webhook.validate(nil, pool)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *InClusterIPPool) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) error {
	newPool, ok := newObj.(types.GenericInClusterPool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a InClusterIPPool or an GlobalInClusterIPPool but got a %T", newObj))
	}
	oldPool, ok := oldObj.(types.GenericInClusterPool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a InClusterIPPool or an GlobalInClusterIPPool but got a %T", oldObj))
	}
	return webhook.validate(oldPool, newPool)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *InClusterIPPool) ValidateDelete(_ context.Context, _ runtime.Object) (reterr error) {
	return nil
}

func (webhook *InClusterIPPool) validate(_, newPool types.GenericInClusterPool) (reterr error) {
	var allErrs field.ErrorList
	defer func() {
		if len(allErrs) > 0 {
			reterr = apierrors.NewInvalid(v1alpha2.GroupVersion.WithKind(newPool.GetObjectKind().GroupVersionKind().Kind).GroupKind(), newPool.GetName(), allErrs)
		}
	}()

	if len(newPool.PoolSpec().Addresses) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "addresses"), newPool.PoolSpec().Addresses, "addresses is required"))
	}

	if newPool.PoolSpec().Prefix == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "prefix"), newPool.PoolSpec().Prefix, "a valid prefix is required"))
	}

	if newPool.PoolSpec().Gateway != "" {
		_, err := netip.ParseAddr(newPool.PoolSpec().Gateway)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "gateway"), newPool.PoolSpec().Gateway, err.Error()))
		}
	}

	for _, address := range newPool.PoolSpec().Addresses {
		if !poolutil.AddressStrParses(address) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "addresses"), address, "provided address is not a valid IP, range, nor CIDR"))
			continue
		}
	}

	if len(allErrs) == 0 {
		errs := validateAddressesAreWithinPrefix(newPool.PoolSpec())
		if len(errs) != 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	return //nolint:nakedret
}

func validatePrefix(spec *v1alpha2.InClusterIPPoolSpec) (*netipx.IPSet, field.ErrorList) {
	var errors field.ErrorList

	addressesIPSet, err := poolutil.AddressesToIPSet(spec.Addresses)
	if err != nil {
		// this should not occur, previous validation should have caught problems here.
		errors := append(errors, field.Invalid(field.NewPath("spec", "addresses"), spec.Addresses, err.Error()))
		return &netipx.IPSet{}, errors
	}

	firstIPInAddresses := addressesIPSet.Ranges()[0].From() // safe because of prior validation
	prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%d", firstIPInAddresses, spec.Prefix))
	if err != nil {
		errors = append(errors, field.Invalid(field.NewPath("spec", "prefix"), spec.Prefix, "provided prefix is not valid"))
		return &netipx.IPSet{}, errors
	}

	builder := netipx.IPSetBuilder{}
	builder.AddPrefix(prefix)
	prefixIPSet, err := builder.IPSet()
	if err != nil {
		// This should not occur, the prefix has been validated. Converting the prefix to an IPSet
		// for it's ContainsRange function.
		errors := append(errors, field.Invalid(field.NewPath("spec", "prefix"), spec.Prefix, err.Error()))
		return &netipx.IPSet{}, errors
	}

	return prefixIPSet, errors
}

func validateAddressesAreWithinPrefix(spec *v1alpha2.InClusterIPPoolSpec) field.ErrorList {
	var errors field.ErrorList

	prefixIPSet, prefixErrs := validatePrefix(spec)
	if len(prefixErrs) > 0 {
		return prefixErrs
	}

	for _, addressStr := range spec.Addresses {
		addressIPSet, err := poolutil.AddressToIPSet(addressStr)
		if err != nil {
			// this should never occur, previous validations will have caught this.
			errors = append(errors, field.Invalid(field.NewPath("spec", "addresses"), addressStr, "provided address is not a valid IP, range, nor CIDR"))
			continue
		}
		// We know that each addressIPSet should be made up of only one range, it came from a single addressStr
		if !prefixIPSet.ContainsRange(addressIPSet.Ranges()[0]) {
			errors = append(errors, field.Invalid(field.NewPath("spec", "addresses"), addressStr, "provided address belongs to a different subnet than others"))
			continue
		}
	}

	return errors
}
