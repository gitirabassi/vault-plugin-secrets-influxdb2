package influxdb2

import (
	"context"
	"errors"
	"net/url"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/mitchellh/mapstructure"
)

const (
	configPath = "config"
)

func (b *backend) pathConfig() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: configPath,
			Fields: map[string]*framework.FieldSchema{
				"host": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "The address of the Influxdb v2 APIs.",
				},
				"token": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "The influxdb token to use.",
				},
				"org_id": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "The Organization ID to use in case the token is not the admin-token of the instance.",
				},
				"initialize": &framework.FieldSchema{
					Type:        framework.TypeBool,
					Description: "Whether the instance still needs to be initialize running the `influx setup` command.",
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.configCreateUpdateOperation,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.configCreateUpdateOperation,
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.configReadOperation,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.configDeleteOperation,
				},
			},
			HelpSynopsis:    configHelpSynopsis,
			HelpDescription: configHelpDescription,
		},
	}
}

func (b *backend) configCreateUpdateOperation(ctx context.Context, req *logical.Request, fieldData *framework.FieldData) (*logical.Response, error) {
	err := fieldData.Validate()
	if err != nil {
		return nil, errors.New("Failing validation: " + err.Error())
	}
	// Host must be provided all the time
	host := fieldData.Get("host").(string)
	if host == "" {
		return nil, errors.New("host is required")
	}
	_, err = url.ParseRequestURI(host)
	if err != nil {
		return nil, errors.New("host is not formatted correctly: " + err.Error())
	}

	initialize := fieldData.Get("initialize").(bool)

	token := fieldData.Get("token").(string)
	if token == "" && !initialize {
		return nil, errors.New("token is required when `initialize` is false or not provided")
	}
	if token != "" && initialize {
		return nil, errors.New("token cannot be set when `initialize` is true")
	}

	orgID := fieldData.Get("org_id").(string)
	if orgID != "" && initialize {
		return nil, errors.New("org_id cannot be configured when `initialize` is true")
	}

	client := influxdb2.NewClient(host, token)
	ok, err := client.Ready(context.Background())

	if err != nil {
		return nil, errors.New("Cannot contact influxdb server: " + err.Error())
	}

	if !ok {
		return nil, errors.New("Influxdb server not Ready: " + err.Error())
	}

	config := &config{
		Host:       host,
		Token:      token,
		OrgID:      orgID,
		Initialize: initialize,
	}

	entry, err := logical.StorageEntryJSON(configPath, config)
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// Respond with a 204.
	return nil, nil
}

func (b *backend) configReadOperation(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	entry, err := req.Storage.Get(ctx, configPath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	config := &config{}
	if err := entry.DecodeJSON(config); err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}
	var configMap map[string]interface{}
	err = mapstructure.Decode(config, &configMap)
	if err != nil {
		return nil, err
	}
	// "token" is intentionally not returned by this endpoint
	delete(configMap, "token")

	resp := &logical.Response{
		Data: configMap,
	}
	return resp, nil
}

func (b *backend) configDeleteOperation(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, configPath); err != nil {
		return nil, err
	}
	return nil, nil
}

type config struct {
	Host       string `json:"host" mapstructure:"host"`
	Token      string `json:"token" mapstructure:"token"`
	OrgID      string `json:"org_id" mapstructure:"org_id"`
	Initialize bool   `json:"initialize" mapstructure:"initialize"`
}

const configHelpSynopsis = `
Configure the Influxdb2 secret engine plugin.
`

const configHelpDescription = `
There are at least 4 scenarios here to be in:
- You have a cloud2 account created and you want to start manage it with Vault. Then provide: host, token, org_id
- You have a OSS Instance of Influxdb v2 managed by someone else and they gave you access to an org. Then provide: host, token, org_id
- You have just startup your own instance of Influxdb 2 OSS, and haven't run 'influx setup' yet. Then provide: host, intitialize=true
- You have just startup your own instance of Influxdb 2 OSS and have already run 'influx setup' but want to start managing as the admin of the instance. Then provide: host, token 
`
