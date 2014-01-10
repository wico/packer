// The cloudstack package contains a packer.Builder implementation
// that builds CloudStack images (templates).

package cloudstack

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mindjiver/gopherstack"
	"log"
	"os"
	"time"
)

// The unique id for the builder
const BuilderId = "mindjiver.cloudstack"

// Configuration tells the builder the credentials to use while
// communicating with CloudStack and describes the template you are
// creating
type config struct {
	common.PackerConfig `mapstructure:",squash"`

	APIURL string `mapstructure:"api_url"`
	APIKey string `mapstructure:"api_key"`
	Secret string `mapstructure:"secret"`

	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	RawStateTimeout string `mapstructure:"state_timeout"`

	SSHUsername string `mapstructure:"ssh_username"`
	SSHPort     uint   `mapstructure:"ssh_port"`
	SSHKeyPath  string `mapstructure:"ssh_key_path"`

	// These are unexported since they're set by other fields
	// being set.
	sshTimeout   time.Duration
	stateTimeout time.Duration

	// Neccessary settings for CloudStack to be able to spin up
	// Virtual Machine.
	ServiceOfferingId string   `mapstructure:"service_offering_id"`
	TemplateId        string   `mapstructure:"template_id"`
	ZoneId            string   `mapstructure:"zone_id"`
	NetworkIds        []string `mapstructure:"network_ids"`

	// Tell CloudStack under which name, description to save the
	// template.
	TemplateName        string `mapstructure:"template_name"`
	TemplateDisplayText string `mapstructure:"template_display_text"`
	TemplateOSId        string `mapstructure:"template_os_id"`

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if b.config.APIURL == "" {
		b.config.APIURL = os.Getenv("CLOUDSTACK_API_URL")
	}

	// Optional configuration with defaults
	if b.config.APIKey == "" {
		// Default to environment variable for api_key, if it exists
		b.config.APIKey = os.Getenv("CLOUDSTACK_API_KEY")
	}

	if b.config.Secret == "" {
		// Default to environment variable for client_id, if it exists
		b.config.Secret = os.Getenv("CLOUDSTACK_SECRET")
	}

	if b.config.ServiceOfferingId == "" {
		b.config.ServiceOfferingId = "62fc8ae5-06ac-4021-bed6-90dfdca6b6b5"
	}

	if b.config.TemplateId == "" {
		b.config.TemplateId = "26de0a07-eee6-4b00-9c4f-fdb7b29f6ba2"
	}

	if b.config.ZoneId == "" {
		b.config.ZoneId = "489e5147-85ba-4f28-a78d-226bf03db47c"
	}

	if len(b.config.NetworkIds) < 1 {
		b.config.NetworkIds = []string{"9ab9719e-1f03-40d1-bfbe-b5dbf598e27f"}
	}

	if b.config.TemplateName == "" {
		// Default to packer-{{ unix timestamp (utc) }}
		b.config.TemplateName = "packer-{{timestamp}}"
	}

	if b.config.TemplateDisplayText == "" {
		b.config.TemplateDisplayText = "Packer Generated Template"
	}

	if b.config.TemplateOSId == "" {
		// Default to Other 64 bit OS
		b.config.TemplateOSId = "103"
	}

	if b.config.SSHUsername == "" {
		// Default to "root". You can override this if your
		// SourceImage has a different user account then the DO default
		b.config.SSHUsername = "root"
	}

	if b.config.SSHPort == 0 {
		// Default to port 22 per DO default
		b.config.SSHPort = 22
	}

	if b.config.RawSSHTimeout == "" {
		// Default to 1 minute timeouts
		b.config.RawSSHTimeout = "1m"
	}

	if b.config.RawStateTimeout == "" {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		b.config.RawStateTimeout = "6m"
	}

	templates := map[string]*string{
		"api_url":       &b.config.APIURL,
		"api_key":       &b.config.APIKey,
		"secret":        &b.config.Secret,
		"template_name": &b.config.TemplateName,
		"ssh_username":  &b.config.SSHUsername,
		"ssh_timeout":   &b.config.RawSSHTimeout,
		"ssh_key_path":  &b.config.SSHKeyPath,
		"state_timeout": &b.config.RawStateTimeout,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	// Required configurations that will display errors if not set
	if b.config.APIURL == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a API URL must be specified"))
	}

	if b.config.APIKey == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("an api_key must be specified"))
	}

	if b.config.Secret == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a secret must be specified"))
	}

	sshTimeout, err := time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	b.config.sshTimeout = sshTimeout

	stateTimeout, err := time.ParseDuration(b.config.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	b.config.stateTimeout = stateTimeout

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	common.ScrubConfig(b.config, b.config.APIKey, b.config.Secret)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Initialize the CloudStack API client
	client := gopherstack.CloudStackClient{}.New(b.config.APIURL, b.config.APIKey, b.config.Secret)

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		new(stepCreateSSHKeyPair),
		new(stepDeployVirtualMachine),
		new(stepVirtualMachineState),
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: 5 * time.Minute,
		},
		new(common.StepProvision),
		new(stepStopVirtualMachine),
		new(stepCreateTemplate),
	}

	// Run the steps
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("template_name"); !ok {
		log.Println("Failed to find template_name in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		templateName: state.Get("template_name").(string),
		templateId:   state.Get("template_id").(string),
		client:       client,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
