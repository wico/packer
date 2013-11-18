package cloudstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateSSHKey struct {
	keyName string
}

func (s *stepCreateSSHKey) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*CloudStackClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary ssh key for virtual machine...")

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	keyName, err := client.CreateSSHKeyPair(name)
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyName = keyName

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_name", keyName)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == 0 {
		return
	}

	client := state.Get("client").(*CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)

	ui.Say("Deleting temporary ssh key...")
	err := client.DeleteSSHKeyPair(s.keyName)

	if err != nil {
		log.Printf("Error cleaning up ssh key: %v", err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually."))
	}
}
