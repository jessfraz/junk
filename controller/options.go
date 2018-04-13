package controller

import (
	"errors"
	"time"
)

// Options holds the options for a controller instance.
type Options struct {
	AzureConfig   string
	KubeConfig    string
	KubeNamespace string

	DomainNameRoot    string
	ResourceGroupName string
	ResourceName      string
	Region            string

	ResyncPeriod time.Duration
}

// validate returns an error if the options are not valid for the controller.
func (opts Options) validate() error {
	if len(opts.AzureConfig) <= 0 {
		return errors.New("Azure config cannot be empty")
	}

	if len(opts.KubeConfig) <= 0 {
		return errors.New("Kube config cannot be empty")
	}

	if len(opts.KubeNamespace) <= 0 {
		return errors.New("Kube namespace cannot be empty")
	}

	if len(opts.DomainNameRoot) <= 0 {
		return errors.New("Domain name root cannot be empty")
	}

	if len(opts.ResourceGroupName) <= 0 {
		return errors.New("Resource group name cannot be empty")
	}

	if len(opts.ResourceName) <= 0 {
		return errors.New("Resource name cannot be empty")
	}

	if len(opts.Region) <= 0 {
		return errors.New("Region cannot be empty")
	}

	return nil
}
