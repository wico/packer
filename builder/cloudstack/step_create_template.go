package cloudstack

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateTemplate struct{}

func (s *stepCreateTemplate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	//	id := state.Get("virtual_machine_id")

	ui.Say(fmt.Sprintf("Creating template: %v", c.TemplateName))

	// get the volume id for the system volume for Virtual Machine 'id'
	volumeid := "0"

	jobId, err := client.CreateTemplate(c.TemplateDisplayText, c.TemplateName, volumeid, c.TemplateOSId)
	if err != nil {
		err := fmt.Errorf("Error creating template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for template to be saved...")
	// Wait for async job?
	err = WaitForAsyncJob(jobId, client, c.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for template to complete: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up template ID for template: %s", c.TemplateName)
	templates, err := client.Templates()
	if err != nil {
		err := fmt.Errorf("Error looking up template ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var templateId string
	for _, template := range templates {
		if template.Name == c.TemplateName {
			templateId = template.Id
			break
		}
	}

	if templateId == "" {
		err := errors.New("Couldn't find template created. Bug?")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Template ID: %d", templateId)

	state.Put("template_id", templateId)
	state.Put("template_name", c.TemplateName)

	return multistep.ActionContinue
}

func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
	// no cleanup
}
