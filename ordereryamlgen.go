package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
)

const _ORDERER_YAML = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  namespace: {{.ns}}
  name: {{.hostname}} 
spec:
  selector:
    matchLabels:
      run: {{.hostname}} 
  replicas: 1
  template:
    metadata:
      labels:
        run: {{.hostname}} 
    spec:
     
      containers:
      - name: {{.hostname}} 
        image: hyperledger/fabric-orderer:x86_64-1.0.0
        env:
          - name: ORDERER_GENERAL_LOGLEVEL
            value: debug
          - name: ORDERER_GENERAL_LISTENADDRESS
            value: "0.0.0.0"
          - name: ORDERER_GENERAL_LISTENPORT
            value: "7050"
          - name: ORDERER_GENERAL_GENESISMETHOD
            value: file
          - name: ORDERER_GENERAL_GENESISFILE
            value: /var/hyperledger/orderer/genesis.block
          - name: ORDERER_GENERAL_LOCALMSPID
            value: OrdererMSP
          - name: ORDERER_GENERAL_LOCALMSPDIR
            value: /var/hyperledger/orderer/msp
          - name: ORDERER_GENERAL_TLS_ENABLED
            value: "true"
          - name: ORDERER_GENERAL_TLS_PRIVATEKEY
            value: /var/hyperledger/orderer/tls/server.key
          - name: ORDERER_GENERAL_TLS_CERTIFICATE
            value: /var/hyperledger/orderer/tls/server.crt
          - name: ORDERER_GENERAL_TLS_ROOTCAS
            value: "[/var/hyperledger/orderer/tls/ca.crt]"          
        command: ["sh","-c","orderer"]
        ports:
        - containerPort: 7050
        volumeMounts:
            - mountPath: /opt
              name: {{.ns}}-nfs
            - mountPath: /var/hyperledger/orderer/genesis.block
              name: {{.ns}}-nfs
              subPath: ws/genesis.block
            - mountPath: /var/hyperledger/orderer/msp
              name: {{.ns}}-nfs
              subPath: ws/crypto-config/ordererOrganizations/{{.ns}}/orderers/{{.hostname}}.{{.ns}}/msp
            - mountPath: /var/hyperledger/orderer/tls
              name: {{.ns}}-nfs
              subPath: ws/crypto-config/ordererOrganizations/{{.ns}}/orderers/{{.hostname}}.{{.ns}}/tls/        
      volumes:
      - name: {{.ns}}-nfs
        persistentVolumeClaim:
          claimName: shared-data
`
const _ORDERER_SERVICE_YAML = `
apiVersion: v1
kind: Service
metadata:
  namespace: {{.ns}}
  name: {{.hostname}}
  labels:
    app: {{.hostname}}
spec:
  ports:
   - port: 7050
     targetPort: 7050
  selector:
     run: {{.hostname}}
    
`

func OrderDeploymentYAMLGen(ordererHostname, ordererDomain string) (string, []string) {
	fileNames := make([]string, 0)
	peerMap := make(map[string]string)
	peerMap["hostname"] = ordererHostname
	peerMap["ns"] = ordererDomain

	tmpl, err := template.New("orderdeployment").Parse(_ORDERER_YAML)
	var outputBytesDep bytes.Buffer
	err = tmpl.Execute(&outputBytesDep, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the orderer deployment yaml file %v\n", err)
		return "", fileNames
	}
	ioutil.WriteFile(ordererHostname+"-deployment.yaml", outputBytesDep.Bytes(), 0666)
	fileNames = append(fileNames, ordererHostname+"-deployment.yaml")
	tmpl, err = template.New("ordererservice").Parse(_ORDERER_SERVICE_YAML)
	var outputBytesService bytes.Buffer
	err = tmpl.Execute(&outputBytesService, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the orderer service yaml file %v\n", err)
		return "", fileNames
	}
	ioutil.WriteFile(ordererHostname+"-service.yaml", outputBytesService.Bytes(), 0666)
	fileNames = append(fileNames, ordererHostname+"-service.yaml")
	return "orderer", fileNames
}
