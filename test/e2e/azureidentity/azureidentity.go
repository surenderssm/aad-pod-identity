package azureidentity

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"os/exec"

	"github.com/Azure/aad-pod-identity/test/e2e/util"
)

// TODO: Add comments
type AzureIdentity struct {
	Metadata Metadata `json:"metadata"`
}

// TODO: Add comments
type Metadata struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
}

// TODO: Add comments
type List struct {
	AzureIdentities []AzureIdentity `json:"items"`
}

// TODO: Add comments
func CreateOnAzure(resourceGroup, name string) error {
	cmd := exec.Command("az", "identity", "-g", resourceGroup, "-n", name)
	util.PrintCommand(cmd)
	_, err := cmd.CombinedOutput()
	return err
}

// TODO: Add comments
func CreateOnCluster(subscriptionID, resourceGroup, name string) error {
	clientID, err := GetClientID(resourceGroup, name)
	if err != nil {
		return err
	}

	t, err := template.New("aadpodidentity.yaml").ParseFiles("deploy/aadpodidentity.yaml")
	if err != nil {
		return err
	}

	deployFile, err := os.Create("deploy/" + name + ".yaml")
	if err != nil {
		return err
	}
	defer deployFile.Close()

	deployData := struct {
		SubscriptionID string
		ResourceGroup  string
		ClientID       string
		Name           string
	}{
		subscriptionID,
		resourceGroup,
		clientID,
		name,
	}
	if err := t.Execute(deployFile, deployData); err != nil {
		return err
	}

	cmd := exec.Command("kubectl", "apply", "-f", "deploy/"+name+".yaml")
	util.PrintCommand(cmd)
	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", err)
	}

	return nil
}

// TODO: Add comments
func GetClientID(resourceGroup, name string) (string, error) {
	cmd := exec.Command("az", "identity", "show", "-g", resourceGroup, "-n", name, "--query", "clientId", "-otsv")
	util.PrintCommand(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// TODO: Add comments
func GetPrincipalID(resourceGroup, name string) (string, error) {
	cmd := exec.Command("az", "identity", "show", "-g", resourceGroup, "-n", name, "--query", "principalId", "-otsv")
	util.PrintCommand(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// TODO: Add comments
func GetAll() (*List, error) {
	cmd := exec.Command("kubectl", "get", "AzureIdentity", "-ojson")
	util.PrintCommand(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error trying to run 'kubectl get AzureIdentity':%s", string(out))
		return nil, err
	}

	nl := List{}
	err = json.Unmarshal(out, &nl)
	if err != nil {
		log.Printf("Error unmarshalling nodes json:%s", err)
	}

	return &nl, nil
}