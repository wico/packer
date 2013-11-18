package cloudstack

import (
	"fmt"
	"log"
)

type Artifact struct {
	// The name of the template
	templateName string

	// The ID of the image
	templateId string

	// The client for making API calls
	client *CloudStackClient
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with CloudStack
	return nil
}

func (a *Artifact) Id() string {
	return a.templateName
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A template was created: %v", a.templateName)
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying template: %d", a.templateId)
	return a.client.DestroyTemplate(a.templateId)
}
