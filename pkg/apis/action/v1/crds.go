package v1

import (
	"fmt"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	wait "k8s.io/apimachinery/pkg/util/wait"
)

type crdConfig struct {
	Kind       string
	Plural     string
	Namespaced bool
	WithStatus bool
}

var myCRDs []*crdConfig = []*crdConfig{
	&crdConfig{"FooBar", "foobars", true, true},
}

func EnsureCRDs(clientset apiextensionsclient.Interface) error {
	for _, crd := range myCRDs {
		err := isCRDExist(clientset, crd)
		if err == nil || apierrors.IsAlreadyExists(err) {
			continue
		}

		err = createCRD(clientset, crd)
		if err != nil {
			return err
		}
	}
	return nil
}

func isCRDExist(clientset apiextensionsclient.Interface, crd *crdConfig) error {
	crdName := fmt.Sprintf("%s.%s", crd.Plural, groupName)
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
	return err
}

func createCRD(clientset apiextensionsclient.Interface, crd *crdConfig) error {
	crdName := fmt.Sprintf("%s.%s", crd.Plural, groupName)
	scope := apiextensionsv1beta1.NamespaceScoped

	if !crd.Namespaced {
		scope = apiextensionsv1beta1.ClusterScoped
	}

	// var subresources *apiextensionsv1beta1.CustomResourceSubresources
	// if crd.WithStatus {
	// 	subresources = &apiextensionsv1beta1.CustomResourceSubresources{
	// 		Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
	// 	}
	// }
	c := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   groupName,
			Version: groupVersion,
			Scope:   scope,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     crd.Plural,
				Kind:       crd.Kind,
				ShortNames: []string{},
			},
			// TODO
			// Subresources: subresources,
		},
	}
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(c)

	// wait for CRD being established
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crdName := fmt.Sprintf("%s.%s", crd.Plural, groupName)
		crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, nil
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					return false, fmt.Errorf("Name conflict: %v", cond.Reason)
				}
			}
		}
		return false, err
	})

	return err
}
