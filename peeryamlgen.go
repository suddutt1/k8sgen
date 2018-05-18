package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
)

const _PEER_DEPLOYMENT_YAML = `
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  namespace: {{ .ns }}
  name: {{ .peerId }}
spec:
  selector:
    matchLabels:
      run: {{ .peerId}}
  replicas: 1
  template:
    metadata:
      labels:
        run: {{ .peerId}}
    spec:
      containers:
        
      - name: {{ .peerId}}{{.org}}
        image: hyperledger/fabric-peer:x86_64-1.0.0
        env:
         - name: CORE_VM_ENDPOINT 
           value: unix:///host/var/run/docker.sock
         - name: CORE_PEER_ADDRESSAUTODETECT
           value: "true"
         - name: CORE_LOGGING_LEVEL 
           value: DEBUG
         - name: CORE_PEER_TLS_ENABLED 
           value: "true" 
         - name: CORE_PEER_ENDORSER_ENABLED 
           value: "true" 
         - name: CORE_PEER_GOSSIP_USELEADERELECTION 
           value: "true" 
         - name: CORE_PEER_GOSSIP_ORGLEADER 
           value: "false"
         - name: CORE_PEER_PROFILE_ENABLED 
           value: "true" 
         - name: CORE_PEER_TLS_CERT_FILE 
           value: /etc/hyperledger/fabric/tls/server.crt
         - name: CORE_PEER_TLS_KEY_FILE 
           value: /etc/hyperledger/fabric/tls/server.key
         - name: CORE_PEER_TLS_ROOTCERT_FILE 
           value: /etc/hyperledger/fabric/tls/ca.crt
         - name: CORE_PEER_ID 
           value: {{ .peerId }}.{{  .ns}}
         - name: CORE_PEER_ADDRESS 
           value: {{ .peerId }}.{{ .ns }}:7051
         - name: CORE_PEER_GOSSIP_EXTERNALENDPOINT 
           value: {{ .peerId }}.{{ .ns }}:7051
         - name: CORE_PEER_LOCALMSPID 
           value: {{ .mspId }}
         - name: CORE_LEDGER_STATE_STATEDATABASE 
           value: goleveldb
         - name: CORE_PEER_GOSSIP_BOOTSTRAP 
           value: {{ .peerId }}.{{ .ns }}:7051
         - name: CORE_VM_DOCKER_ATTACHSTDOUT
           value: "true"
         - name: CORE_PEER_TLS_SERVERHOSTOVERRIDE
           value: {{ .peerId }}.{{ .ns }}
         - name: CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE
           value: bridge
        command: ["sh", "-c", "peer node start"]
        ports:
        - containerPort: 7051
        - containerPort: 7052
        - containerPort: 7053
        volumeMounts:
          - mountPath: /opt/ws
            name: {{.ns}}-{{.org}}-{{.peerId}}-nfs
            subPath: ws
          - mountPath: /etc/hyperledger/fabric/msp
            name: {{.ns}}-{{.org}}-{{.peerId}}-nfs
            subPath: ws/crypto-config/peerOrganizations/{{ .ns }}/peers/{{ .peerId }}.{{ .ns }}/msp
          - mountPath: /etc/hyperledger/fabric/tls
            name: {{.ns}}-{{.org}}-{{.peerId}}-nfs
            subPath: ws/crypto-config/peerOrganizations/{{ .ns }}/peers/{{ .peerId }}.{{ .ns }}/tls
          - mountPath: /host/var/run/docker.sock
            name: dockersocket
         
      volumes:
      - name: {{.ns}}-{{.org}}-{{.peerId}}-nfs
        persistentVolumeClaim:
          claimName: shared-data
      - name: dockersocket
        hostPath:
          path: /var/run/docker.sock

        
---

`

const _PEER_SERVICE_YAML = `
apiVersion: v1
kind: Service
metadata:
   namespace: {{.ns}}
   name: {{.peerId}}
   labels:
    app: {{.peerId}}
spec:
   ports:
      - name: externale-listen-endpoint
        protocol: TCP
        port: 7051
        targetPort: 7051
      - name: chaincode-listen
        protocol: TCP
        port: 7052
        targetPort: 7052
      - name: event-listen
        protocol: TCP
        port: 7053
        targetPort: 7053 
   selector:
        run: {{.peerId}}

---
`
const _NAMESPACE_YAML = `
apiVersion: v1
kind: Namespace
metadata:
  name: {{.ns}}
`

func NameSpaceYAMLGen(namespace string) (string, []string) {
	fileNames := make([]string, 0)
	peerMap := make(map[string]string)

	peerMap["ns"] = namespace

	tmpl, err := template.New("nsdeployment").Parse(_NAMESPACE_YAML)
	var outputBytesDep bytes.Buffer
	err = tmpl.Execute(&outputBytesDep, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the peer namespace yaml file %v\n", err)
		return "", fileNames
	}
	ioutil.WriteFile(namespace+"-ns.yaml", outputBytesDep.Bytes(), 0666)
	fileNames = append(fileNames, namespace+"-ns.yaml")
	return "node-ns", fileNames

}
func PeerDeploymentYAMLGen(peerName, peerDomain, mspId, org string) (string, []string) {
	fileNames := make([]string, 0)
	peerMap := make(map[string]string)
	peerMap["peerId"] = peerName
	peerMap["ns"] = peerDomain
	peerMap["mspId"] = mspId
	peerMap["org"] = org
	tmpl, err := template.New("peerdeployment").Parse(_PEER_DEPLOYMENT_YAML)
	var outputBytesDep bytes.Buffer
	err = tmpl.Execute(&outputBytesDep, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the peer deployment yaml file %v\n", err)
		return "", fileNames
	}
	ioutil.WriteFile(org+"-"+peerName+"-deployment.yaml", outputBytesDep.Bytes(), 0666)
	fileNames = append(fileNames, org+"-"+peerName+"-deployment.yaml")
	tmpl, err = template.New("peerservice").Parse(_PEER_SERVICE_YAML)
	var outputBytesService bytes.Buffer
	err = tmpl.Execute(&outputBytesService, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the peer service yaml file %v\n", err)
		return "", fileNames
	}
	ioutil.WriteFile(org+"-"+peerName+"-service.yaml", outputBytesService.Bytes(), 0666)
	fileNames = append(fileNames, org+"-"+peerName+"-service.yaml")
	return "peer", fileNames
}
