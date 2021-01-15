/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MysqlPhase string

const (
	Success     MysqlPhase = "SUCCESS"
	Available   MysqlPhase = "AVAILABLE"
	Unavailable MysqlPhase = "UNAVAILABLE"
	Pending     MysqlPhase = "PENDING"
	Fail        MysqlPhase = "FAIL"
)

// MysqlHAClusterSpec defines the desired state of MysqlHACluster
type MysqlHAClusterSpec struct {
	// backUpQuantity: the quantity of mysql backup server
	BackUpQuantity uint `json:"backUpQuantity,omitempty"`
	// followerQuantity: the quantity of mysql follower server
	FollowerQuantity uint `json:"followerQuantity,omitempty"`
	// A label query over pods that are managed by the mysql cluster
	Selector map[string]string `json:"selector,omitempty" protobuf:"bytes,1,rep,name=selector"`
	// MysqlClientImage: image of mysql client
	MysqlClientImage string `json:"mysqlClientImage,omitempty"`
	// MysqlServer: mysql server pod template
	MysqlServer v1.PodTemplateSpec `json:"mysqlServer" protobuf:"bytes,6,opt,name=mysqlServer"`
	// MysqlProxy: mysql proxy pod template
	MysqlProxy v1.PodTemplateSpec `json:"mysqlProxy" protobuf:"bytes,6,opt,name=mysqlProxy"`
	// UserName: mysql connect username
	UserName string `json:"username,omitempty"`
	// Password: mysql connect password
	Password string `json:"password,omitempty"`
	// ExposePort: mysql connect port out of cluster
	ExposePort uint8 `json:"exposePort,omitempty"`
}

// MysqlHAClusterMasterInfo defines the desired master info of MysqlHACluster
type MysqlHAClusterMasterInfo struct {
	//podIp: the ip of the master pod
	PodIp string `json:"podIp,omitempty"`
	// podName: the name of the master pod
	PodName string `json:"podName,omitempty"`
	// nodeName: the master pod location node name
	NodeName string `json:"nodeName,omitempty"`
	// nodeIp: the master pod location node ip
	NodeIp string `json:"nodeIp,omitempty"`
}

// MysqlHAClusterStatus defines the observed state of MysqlHACluster
type MysqlHAClusterStatus struct {
	Phase          MysqlPhase               `json:"phase,omitempty"`
	BackUpExpect   uint                     `json:"backUpExpect,omitempty"`
	BackUpReady    uint                     `json:"backUpReady,omitempty"`
	FollowerExpect uint                     `json:"followerExpect,omitempty"`
	FollowerReady  uint                     `json:"followerReady,omitempty"`
	Master         MysqlHAClusterMasterInfo `json:"master,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.phase",name=status,type=string
//+kubebuilder:printcolumn:JSONPath=".status.backUpExpect",name=backUpExpect,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.backUpReady",name=backUpReady,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.followerExpect",name=followerExpect,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.followerReady",name=followerReady,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.master.podIp",name=masterIp,type=string
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Name",type="string",JSONPath=".metadata.name"
// MysqlHACluster is the Schema for the mysqlhaclusters API
type MysqlHACluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlHAClusterSpec   `json:"spec,omitempty"`
	Status MysqlHAClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlHAClusterList contains a list of MysqlHACluster
type MysqlHAClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlHACluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlHACluster{}, &MysqlHAClusterList{})
}
