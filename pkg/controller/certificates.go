/*
Copyright 2021 Jetstack Ltd.
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

package controller

import (
	"encoding/json"
	"fmt"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
)

type Message struct {
	Operation  string `json:"operation"`
	CertSpec   cmapi.CertificateSpec `json:"cert_spec"`
	CertStatus cmapi.CertificateStatus `json:"cert_status"`
}

func (c *Context) addCertificate(obj interface{}) {
	cert := obj.(*cmapi.Certificate)
	c.Log.V(5).Info(cert.ObjectMeta.Namespace, cert.ObjectMeta.Namespace, "created")
	subject := fmt.Sprintf("io.cert-manager.certificates.%s.%s", cert.GetNamespace(), cert.GetName())
	msg, err := json.Marshal(Message{
		Operation:  "add",
		CertSpec:   cert.Spec,
		CertStatus: cert.Status,
	})
	if err != nil {
		c.Log.Error(err, "couldn't publish message")
	}
	if err := c.NatsClient.Publish(subject, msg); err != nil {
		c.Log.Error(err, "couldn't publish message")
	}
}

func (c *Context) updateCertificate(old, new interface{}) {
	oldCert := old.(*cmapi.Certificate)
	cert := new.(*cmapi.Certificate)
	c.Log.V(5).Info(oldCert.ObjectMeta.Namespace, oldCert.ObjectMeta.Namespace, "updated")
	subject := fmt.Sprintf("io.cert-manager.certificates.%s.%s", cert.GetNamespace(), cert.GetName())
	msg, err := json.Marshal(Message{
		Operation:  "update",
		CertSpec:   cert.Spec,
		CertStatus: cert.Status,
	})
	if err != nil {
		c.Log.Error(err, "couldn't publish message")
	}
	if err := c.NatsClient.Publish(subject, msg); err != nil {
		c.Log.Error(err, "couldn't publish message")
	}
}

func (c *Context) deleteCertificate(obj interface{}) {
	cert := obj.(*cmapi.Certificate)
	c.Log.V(5).Info(cert.ObjectMeta.Namespace, cert.ObjectMeta.Namespace, "deleted")
	subject := fmt.Sprintf("io.cert-manager.certificates.%s.%s", cert.GetNamespace(), cert.GetName())
	msg, err := json.Marshal(Message{
		Operation:  "delete",
		CertSpec:   cert.Spec,
		CertStatus: cert.Status,
	})
	if err != nil {
		c.Log.Error(err, "couldn't publish message")
	}
	if err := c.NatsClient.Publish(subject, msg); err != nil {
		c.Log.Error(err, "couldn't publish message")
	}
}
