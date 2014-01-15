package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepDetachIso struct{}

func (s *stepDetachIso) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudStackClient)
	c := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("virtual_machine_id").(string)

	// Stop the virtual machine
	ui.Say("Detaching ISO image...")
	jobId, err := client.DetachIso(id)
	if err != nil {
		err := fmt.Errorf("Error detaching ISO from virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for detach event to complete...")
	err = client.WaitForAsyncJob(jobId, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDetachIso) Cleanup(state multistep.StateBag) {
	// no cleanup
}
