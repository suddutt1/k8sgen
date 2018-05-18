package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func GenerateYamlFiles(config []byte) {
	cmdInput := make(map[string][]string)
	networkConfig := make(map[string]interface{})
	json.Unmarshal(config, &networkConfig)
	ordererConfig := getMap(networkConfig["orderers"])
	ordererDomain := getString(ordererConfig["domain"])
	ordererHostname := getString(ordererConfig["ordererHostname"])
	id, files := OrderDeploymentYAMLGen(ordererHostname, ordererDomain)
	cmdInput[id] = files
	orgConfigs, _ := networkConfig["orgs"].([]interface{})
	for index, org := range orgConfigs {
		orgConfig, _ := org.(map[string]interface{})
		fmt.Printf("Processing org %d \n", index)
		peerCountFlt, _ := orgConfig["peerCount"].(float64)
		peerCount := int(peerCountFlt)
		fmt.Printf(" Peer count is %d \n ", peerCount)
		peerDomain := getString(orgConfig["domain"])
		orgName := strings.ToLower(getString(orgConfig["name"]))
		for peerIndex := 0; peerIndex < peerCount; peerIndex++ {
			peerID := fmt.Sprintf("peer%d", peerIndex)

			id, files = PeerDeploymentYAMLGen(peerID, peerDomain, getString(orgConfig["mspID"]), orgName)
			cmdInput[id+orgName] = files
		}
		id, files = NameSpaceYAMLGen(peerDomain)
		cmdInput[id+orgName] = files
		id, files = VolumeClaimForNamespace(peerDomain)
		cmdInput[id+orgName] = files
	}
	id, files = NameSpaceYAMLGen(ordererDomain)
	cmdInput[id+ordererHostname] = files
	id, files = VolumeYamlGenerator(ordererDomain)
	cmdInput[id] = files
	id, files = VolumeClaimForNamespace(ordererDomain)
	cmdInput[id+ordererHostname] = files
	id, files = NoopCliYAMLGen(ordererDomain)
	cmdInput[id] = files
	id, files = HLFCliYAMLGen(ordererDomain)
	cmdInput[id] = files
	printInOrder(cmdInput)
}
func printInOrder(filesGenerated map[string][]string) {
	printCommands("nfs", filesGenerated)
	printCommands("node-ns", filesGenerated)

	printCommands("node-pvc", filesGenerated)
	fmt.Println("Update IP for the following...")
	fmt.Println("export NFS_IP=`kubectl get service | grep -e \"nfs-server\"| awk '{print $3}'`")
	fmt.Println("sed -i \"s/__IP_ADDRESS__/${NFS_IP}/g\" *-nfs-pv.yaml")
	fmt.Println("NOOP CLI... Create the deployables")
	printCommands("noop", filesGenerated)
	printCommands("orderer", filesGenerated)
	printCommands("peer", filesGenerated)
	printCommands("hlf", filesGenerated)

}
func printCommands(key string, filesGenerated map[string][]string) {
	fmt.Println("-------------------------------------------")
	for k, files := range filesGenerated {
		if strings.HasPrefix(k, key) {
			for _, fileName := range files {
				fmt.Printf("kubectl apply -f ./%s\n", fileName)
			}
		}
	}

}
