package cloudstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"github.com/mindjiver/gopherstack"
	"log"
)

type stepCreateSSHKeyPair struct {
	keyName    string
	privateKey string
}

func (s *stepCreateSSHKeyPair) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary ssh key for virtual machine...")

	// The name of the public key on CloudStack
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	privateKey, err := client.CreateSSHKeyPair(name)
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyName = name
	s.privateKey = privateKey

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_name", name)
	state.Put("ssh_private_key", privateKey)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKeyPair) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary ssh key...")
	_, err := client.DeleteSSHKeyPair(s.keyName)

	if err != nil {
		log.Printf("Error cleaning up ssh key: %v", err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually."))
	}
}
