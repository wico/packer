package cloudstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepVirtualMachineState struct{}

func (s *stepVirtualMachineState) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	id := state.Get("virtual_machine_id")

	ui.Say("Waiting for virtual machine to become active...")

	err := waitForDropletState("active", dropletId, client, c.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become active: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	ip, _, err := client.DropletStatus(dropletId)
	if err != nil {
		err := fmt.Errorf("Error retrieving virtual machine IP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("virtual_machine_ip", ip)

	return multistep.ActionContinue
}

func (s *stepVirtualMachineState) Cleanup(state multistep.StateBag) {
	// no cleanup
}
