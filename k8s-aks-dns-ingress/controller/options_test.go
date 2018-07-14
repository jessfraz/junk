package controller

import (
	"testing"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"k8s.io/client-go/kubernetes/fake"
)

func TestValidateOptionsAllEmpty(t *testing.T) {
	opts := Options{}
	err := opts.validate()

	if err != errAzureAuthenticationNil {
		t.Fatalf("expected error %v, got %v", errAzureAuthenticationNil, err)
	}
}

func TestValidateOptionsKubeClientEmpty(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
	}
	err := opts.validate()

	if err != errKubeClientNil {
		t.Fatalf("expected error %v, got %v", errKubeClientNil, err)
	}
}

func TestValidateOptionsDomainNameRootEmpty(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		KubeClient: fake.NewSimpleClientset(),
	}
	err := opts.validate()

	if err != errDomainNameRootEmpty {
		t.Fatalf("expected error %v, got %v", errDomainNameRootEmpty, err)
	}
}

func TestValidateOptionsResourceGroupNameEmpty(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		KubeClient:     fake.NewSimpleClientset(),
		DomainNameRoot: fakeDomainNameRoot,
	}
	err := opts.validate()

	if err != errResourceGroupNameEmpty {
		t.Fatalf("expected error %v, got %v", errResourceGroupNameEmpty, err)
	}
}

func TestValidateOptionsResourceNameEmpty(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		KubeClient:        fake.NewSimpleClientset(),
		DomainNameRoot:    fakeDomainNameRoot,
		ResourceGroupName: fakeResourceGroupName,
	}
	err := opts.validate()

	if err != errResourceNameEmpty {
		t.Fatalf("expected error %v, got %v", errResourceNameEmpty, err)
	}
}

func TestValidateOptionsRegionEmpty(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		KubeClient:        fake.NewSimpleClientset(),
		DomainNameRoot:    fakeDomainNameRoot,
		ResourceGroupName: fakeResourceGroupName,
		ResourceName:      fakeResourceName,
	}
	err := opts.validate()

	if err != errRegionEmpty {
		t.Fatalf("expected error %v, got %v", errRegionEmpty, err)
	}
}

func TestValidateOptionsResyncPeriodZero(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		KubeClient:        fake.NewSimpleClientset(),
		DomainNameRoot:    fakeDomainNameRoot,
		ResourceGroupName: fakeResourceGroupName,
		ResourceName:      fakeResourceName,
		Region:            fakeRegion,
	}
	if err := opts.validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateOptionsOK(t *testing.T) {
	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		KubeClient:        fake.NewSimpleClientset(),
		DomainNameRoot:    fakeDomainNameRoot,
		ResourceGroupName: fakeResourceGroupName,
		ResourceName:      fakeResourceName,
		Region:            fakeRegion,
		ResyncPeriod:      time.Second * 30,
	}
	if err := opts.validate(); err != nil {
		t.Fatal(err)
	}
}
