package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
)

const _NOOP_CLI_YAML = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  namespace: {{.ns}}
  name: noopcli
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: noopcli
    spec:
      containers:
      - name: noopcli
        image: suddutt1/noopcli:2.0.0
        volumeMounts:
            # name must match the volume name below
            - name: {{.ns}}-nfs
              mountPath: "/opt"
      volumes:
      - name: {{.ns}}-nfs
        persistentVolumeClaim:
          claimName: shared-data

`

func NoopCliYAMLGen(namespace string) (string, []string) {
	fileName := make([]string, 1)
	peerMap := make(map[string]string)
	peerMap["ns"] = namespace
	tmpl, err := template.New("noopcli").Parse(_NOOP_CLI_YAML)
	var outputBytesDep bytes.Buffer
	err = tmpl.Execute(&outputBytesDep, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the noopcli deployment yaml file %v\n", err)
		return "noopcli", fileName
	}
	ioutil.WriteFile("noopcli-deployment.yaml", outputBytesDep.Bytes(), 0666)
	fileName[0] = "noopcli-deployment.yaml"
	return "noopcli", fileName
}

const _HLFCLI_YAML = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  namespace: {{.ns}}
  name: hlfcli
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: hlfcli
    spec:
     
      containers:
      - name: hlfcli
        image: hyperledger/fabric-tools:x86_64-1.0.0
        env:
          - name: CORE_PEER_TLS_ENABLED
            value: "true"
          - name: GOPATH
            value: /opt/gopath
          - name: CORE_LOGGING_LEVEL
            value: DEBUG
          - name: CORE_PEER_ID
            value: cli
        command: ["sh","-c","sleep 9999999"]
        volumeMounts:
            - mountPath: /opt/ws/
              name: nfs
              subPath: ws
            - mountPath: /opt/gopath/src
              name: nfs
              subPath: ws/chaincode/  

      volumes:
      - name: nfs
        persistentVolumeClaim:
          claimName:  shared-data

`

func HLFCliYAMLGen(namespace string) (string, []string) {
	fileNames := make([]string, 0)
	peerMap := make(map[string]string)
	peerMap["ns"] = namespace
	tmpl, err := template.New("hlfcli").Parse(_HLFCLI_YAML)
	var outputBytesDep bytes.Buffer
	err = tmpl.Execute(&outputBytesDep, peerMap)
	if err != nil {
		fmt.Printf("Error in generating the hlfcli deployment yaml file %v\n", err)
		return "", fileNames
	}
	ioutil.WriteFile("hlfcli-deployment.yaml", outputBytesDep.Bytes(), 0666)
	fileNames = append(fileNames, "hlfcli-deployment.yaml")
	return "hlfcli", fileNames
}

const _SSD_STORAGE_YAML = `

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ssd-disk
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd

`
const _SSD_VC_YAML = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ssd-volume
spec:
  storageClassName: ssd-disk
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 2Gi

`
const _NFS_SERVER_YAML = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: nfs-server
spec:
  replicas: 1
  selector:
    role: nfs-server
  template:
    metadata:
      labels:
        role: nfs-server
    spec:
      containers:
      - name: nfs-server
        image: k8s.gcr.io/volume-nfs:0.8
        ports:
          - name: nfs
            containerPort: 2049
          - name: mountd
            containerPort: 20048
          - name: rpcbind
            containerPort: 111
        securityContext:
          privileged: true
        volumeMounts:
          - mountPath: /exports
            name: mypvc
      volumes:
        - name: mypvc
          persistentVolumeClaim:
            claimName: ssd-volume

`
const _NFS_SERVICE_YAML = `
kind: Service
apiVersion: v1
metadata:
  name: nfs-server
spec:
  ports:
    - name: nfs
      port: 2049
    - name: mountd
      port: 20048
    - name: rpcbind
      port: 111
  selector:
      role: nfs-server
`
const _NFS_PV_YAML = `
apiVersion: v1
kind: PersistentVolume
metadata:
  namespace: {{.ns}}
  name: {{.ns}}-nfs
spec:
  capacity:
    storage: 256Mi
  accessModes:
    - ReadWriteMany
  nfs:
    # FIXME: use the right IP
    server: __IP_ADDRESS__
    path: "/"
    
---  
`
const _NFS_PVC_YAML = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  namespace: {{.ns}}
  name: shared-data
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: ""
  resources:
    requests:
      storage: 200Mi
`

func VolumeYamlGenerator(domainName string) (string, []string) {
	fileNames := make([]string, 0)
	dataMap := make(map[string]string)
	dataMap["ns"] = domainName
	generateFiles(_SSD_STORAGE_YAML, "ssd-storage-class.yaml", dataMap)
	fileNames = append(fileNames, "ssd-storage-class.yaml")
	generateFiles(_SSD_VC_YAML, "ssd-volume-claim.yaml", dataMap)
	fileNames = append(fileNames, "ssd-volume-claim.yaml")
	generateFiles(_NFS_SERVER_YAML, "nfs-server.yaml", dataMap)
	fileNames = append(fileNames, "nfs-server.yaml")
	generateFiles(_NFS_SERVICE_YAML, "nfs-service.yaml", dataMap)
	fileNames = append(fileNames, "nfs-service.yaml")
	return "nfs", fileNames
}
func VolumeClaimForNamespace(domainName string) (string, []string) {
	fileNames := make([]string, 0)
	dataMap := make(map[string]string)
	dataMap["ns"] = domainName
	generateFiles(_NFS_PV_YAML, domainName+"-nfs-pv.yaml", dataMap)
	fileNames = append(fileNames, domainName+"-nfs-pv.yaml")
	generateFiles(_NFS_PVC_YAML, domainName+"-nfs-pvc.yaml", dataMap)
	fileNames = append(fileNames, domainName+"-nfs-pvc.yaml")
	return "node-pvc", fileNames
}
func generateFiles(templateName, outputName string, data map[string]string) {
	tmpl, err := template.New("templ" + templateName).Parse(templateName)
	var outputBytesDep bytes.Buffer
	err = tmpl.Execute(&outputBytesDep, data)
	if err != nil {
		fmt.Printf("Error in generating the %s file %v\n", outputName, err)
		return
	}
	ioutil.WriteFile(outputName, outputBytesDep.Bytes(), 0666)
}
