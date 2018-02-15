package network

import (
	"context"

	"github.com/docker/go-plugins-helpers/ipam"
)

type IPAMDriver struct {
	ctx context.Context
}

func NewIPAMDriver(ctx context.Context) *IPAMDriver {
	return &IPAMDriver{
		ctx: ctx,
	}
}

func (d *IPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (d *IPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {

	return &ipam.AddressSpacesResponse{}, nil
}

func (d *IPAMDriver) RequestPool(*ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {

	return nil, nil
}

func (d *IPAMDriver) ReleasePool(*ipam.ReleasePoolRequest) error {

	return nil
}

func (d *IPAMDriver) RequestAddress(*ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {

	return nil, nil
}

func (d *IPAMDriver) ReleaseAddress(*ipam.ReleaseAddressRequest) error {

	return nil
}
