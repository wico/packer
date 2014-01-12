package cloudstack

import (
	"errors"
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateTemplate struct{}

func (s *stepCreateTemplate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	vmid := state.Get("virtual_machine_id").(string)

	ui.Say(fmt.Sprintf("Creating template: %v", c.TemplateName))

	// get the volume id for the system volume for Virtual Machine 'id'
	volumeId, err := client.ListVolumes(vmid)
	if err != nil {
		err := fmt.Errorf("Error creating template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	jobId, err := client.CreateTemplate(c.TemplateDisplayText, c.TemplateName,
		volumeId, c.TemplateOSId)
	if err != nil {
		err := fmt.Errorf("Error creating template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for template to be saved...")
	err = client.WaitForAsyncJob(jobId, c.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for template to complete: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up template ID for template: %s", c.TemplateName)
	template, templateId, err := client.ListTemplates(c.TemplateName, "self")
	if err != nil {
		err := fmt.Errorf("Error looking up template ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if template != c.TemplateName {
		err := errors.New("Couldn't find template created. Bug?")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("template_name", template)
	state.Put("template_id", templateId)

	return multistep.ActionContinue
}

func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
	// no cleanup
}
