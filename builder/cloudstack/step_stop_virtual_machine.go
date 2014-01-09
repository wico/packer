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
	id := state.Get("virtual_machine_id").(string)

	_, status, err := client.VirtualMachineState(id)
	if err != nil {
		err := fmt.Errorf("Error checking virtual machine state: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if status == "stopped" {
		// Virtual Machine is already stopped, don't do anything
		return multistep.ActionContinue
	}

	// Stop the virtual machine
	ui.Say("Stopping Virtual Machine...")
	jobId, err := client.StopVirtualMachine(id)
	if err != nil {
		err := fmt.Errorf("Error stopping virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for stop event to complete...")
	err = WaitForAsyncJob(jobId, client, c.stateTimeout)
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
