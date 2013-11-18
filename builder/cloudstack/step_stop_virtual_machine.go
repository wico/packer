package cloudstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepStopVirtualMachine struct{}

func (s *stepStopVirtualMachine) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*CloudStackClient)
	c := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("virtual_machine_id").(uint)

	_, status, err := client.VirtualMachineState(dropletId)
	if err != nil {
		err := fmt.Errorf("Error checking virtual machine state: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if status == "off" {
		// Droplet is already off, don't do anything
		return multistep.ActionContinue
	}

	// Stop the virtual machine
	ui.Say("Stopping Virtual Machine...")
	err = client.StopVirtualMachine(id)
	if err != nil {
		err := fmt.Errorf("Error powering off virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// FIXME: Implement this function
	log.Println("Waiting for stop event to complete...")
	err = waitForDropletState("off", id, client, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopVirtualMachine) Cleanup(state multistep.StateBag) {
	// no cleanup
}
