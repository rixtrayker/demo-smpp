package app

import (
	"github.com/rixtrayker/demo-smpp/internal/smpp"
)

func HandleProviderA(smppClient *smpp.Client, message string) error {
	// handle sending message for provider A
	return smppClient.SendMessage("A", message)
}

func HandleProviderB(smppClient *smpp.Client, message string) error {
	// handle sending message for provider B
	return smppClient.SendMessage("B", message)
}

func HandleProviderC(smppClient *smpp.Client, message string) error {
	// handle sending message for provider C
	return smppClient.SendMessage("C", message)
}
