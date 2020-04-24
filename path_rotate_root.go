package influxdb2

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	rotateRootPath = "config/rotate-root"
)

func (b *backend) pathRotateRoot() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: rotateRootPath,
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRotateRootWrite,
				},
			},
			HelpSynopsis:    pathRotateRootReadHelpSyn,
			HelpDescription: pathRotateRootReadHelpDesc,
		},
	}
}

func (b *backend) pathRotateRootWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.Lock()
	defer b.Unlock()

	configEntry, err := req.Storage.Get(ctx, configPath)
	if err != nil {
		return logical.ErrorResponse("influxdb config is missing (get): %v", err), nil
	}
	if configEntry == nil {
		return logical.ErrorResponse("influxdb config is missing (nil response): %v", err), nil
	}
	config := &config{}
	if err := configEntry.DecodeJSON(config); err != nil {
		return logical.ErrorResponse("influxdb config is missing (decode json): %v", err), nil
	}
	if config == nil {
		return logical.ErrorResponse("influxdb config is missing (decoded json is nil): %v", err), nil
	}
	// TODO: actually rotate ROot credential
	return nil, nil
}

const pathRotateRootReadHelpSyn = `
Request Influxdb credentials for a certain role. These credentials are
rotated periodically.`

const pathRotateRootReadHelpDesc = `
This path reads influxdb v2 credentials for a certain role. 
`
