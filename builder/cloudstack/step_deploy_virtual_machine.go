package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"time"
)

type stepDeployVirtualMachine struct {
	id string
}

func (s *stepDeployVirtualMachine) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	sshKeyName := state.Get("ssh_key_name").(string)

	ui.Say("Creating virtual machine...")

	// Some random virtual machine name as it's temporary
	displayName := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the virtual machine based on configuration
	vmid, jobid, err := client.DeployVirtualMachine(c.ServiceOfferingId, c.TemplateId, c.ZoneId, c.NetworkIds, sshKeyName, displayName, "", "")
	if err != nil {
		err := fmt.Errorf("Error deploying Virtual Machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	client.WaitForAsyncJob(jobid, 2*time.Minute)
	// TODO: add error handling here

	// We use this in cleanup
	s.id = vmid

	// Store the virtual machine id for later use
	state.Put("virtual_machine_id", vmid)
	//state.Put("root_device_id",
	return multistep.ActionContinue
}

func (s *stepDeployVirtualMachine) Cleanup(state multistep.StateBag) {
	// If the virtual machine id isn't there, we probably never created it
	if s.id == "" {
		return
	}

	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)

	// Destroy the droplet we just created
	ui.Say("Destroying virtual machine...")

	jobid, err := client.DestroyVirtualMachine(s.id)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually."))
	}

	client.WaitForAsyncJob(jobid, 2*time.Minute)
	// TODO: add error handling here
}
