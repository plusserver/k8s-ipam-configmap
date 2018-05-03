package main

import (
	"testing"

	"github.com/Nexinto/go-ipam"
	"github.com/Nexinto/k8s-ipam-shared"
	"github.com/stretchr/testify/assert"

	"k8s.io/client-go/kubernetes/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ipamv1 "github.com/Nexinto/k8s-ipam/pkg/apis/ipam.nexinto.com/v1"
	fakeipamclientset "github.com/Nexinto/k8s-ipam/pkg/client/clientset/versioned/fake"
)

func GetTestController() *Controller {
	kubernetes := fake.NewSimpleClientset()
	ipamclient := fakeipamclientset.NewSimpleClientset()
	tmpl, _ := MakeNameTemplate()
	ipams, _ := ipam.NewConfigMapIpam(kubernetes, "10.10.0.0/16")

	kubernetes.CoreV1().Namespaces().Create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}})

	return &Controller{
		Kubernetes: kubernetes,
		IpamClient: ipamclient,
		SharedController: ipamshared.SharedController{
			Kubernetes:   kubernetes,
			IpamClient:   ipamclient,
			Ipam:         ipams,
			Tag:          "fakecm",
			NameTemplate: tmpl,
			IpamName:     "ConfigMap",
		},
	}
}

func TestFreeAddress(t *testing.T) {
	a := assert.New(t)

	c := GetTestController()

	ia := &ipamv1.IpAddress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "myip",
		},
	}

	_, err := c.IpamClient.IpamV1().IpAddresses("default").Create(ia)
	if !a.Nil(err) {
		return
	}

	err = c.IpAddressCreatedOrUpdated(ia)
	if !a.Nil(err) {
		return
	}

	ia, err = c.IpamClient.IpamV1().IpAddresses("default").Get("myip", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}

	a.Equal("10.10.0.0", ia.Status.Address)
	a.Equal("fakecm.default.myip", ia.Status.Name)

	cm, err := c.Kubernetes.CoreV1().ConfigMaps("kube-system").Get("ipam-cm", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}

	if !a.NotEmpty(cm.Data) {
		return
	}
	if !a.NotEmpty(cm.Data["10.10.0.0"]) {
		return
	}
	a.Equal("fakecm.default.myip", cm.Data["10.10.0.0"])

	err = c.IpamClient.IpamV1().IpAddresses("default").Delete("myip", &metav1.DeleteOptions{})
	if !a.Nil(err) {
		return
	}

	err = c.IpAddressDeleted(ia)
	if !a.Nil(err) {
		return
	}

	cm, err = c.Kubernetes.CoreV1().ConfigMaps("kube-system").Get("ipam-cm", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}
	if !a.Empty(cm.Data) {
		return
	}
	a.Empty(cm.Data["10.10.0.0"])

}

func TestReference(t *testing.T) {
	a := assert.New(t)

	c := GetTestController()

	ia := &ipamv1.IpAddress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "myip",
		},
	}

	_, err := c.IpamClient.IpamV1().IpAddresses("default").Create(ia)
	if !a.Nil(err) {
		return
	}

	err = c.IpAddressCreatedOrUpdated(ia)
	if !a.Nil(err) {
		return
	}

	ia, err = c.IpamClient.IpamV1().IpAddresses("default").Get("myip", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}

	ir := &ipamv1.IpAddress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "myref",
		},
		Spec: ipamv1.IpAddressSpec{
			Ref: "fakecm.default.myip",
		},
	}

	_, err = c.IpamClient.IpamV1().IpAddresses("default").Create(ir)
	if !a.Nil(err) {
		return
	}

	err = c.IpAddressCreatedOrUpdated(ir)
	if !a.Nil(err) {
		return
	}

	ir, err = c.IpamClient.IpamV1().IpAddresses("default").Get("myref", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}

	a.Equal(ir.Status.Address, ia.Status.Address)

	cm, err := c.Kubernetes.CoreV1().ConfigMaps("kube-system").Get("ipam-cm", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}

	if !a.NotEmpty(cm.Data) {
		return
	}
	if !a.NotEmpty(cm.Data["10.10.0.0"]) {
		return
	}
	a.Equal("fakecm.default.myip", cm.Data["10.10.0.0"])

	err = c.IpamClient.IpamV1().IpAddresses("default").Delete("myref", &metav1.DeleteOptions{})
	if !a.Nil(err) {
		return
	}

	err = c.IpAddressDeleted(ir)
	if !a.Nil(err) {
		return
	}

	cm, err = c.Kubernetes.CoreV1().ConfigMaps("kube-system").Get("ipam-cm", metav1.GetOptions{})
	if !a.Nil(err) {
		return
	}

	if !a.NotEmpty(cm.Data) {
		return
	}
	if !a.NotEmpty(cm.Data["10.10.0.0"]) {
		return
	}
	a.Equal("fakecm.default.myip", cm.Data["10.10.0.0"])
}
