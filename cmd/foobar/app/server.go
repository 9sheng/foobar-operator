package app

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	options "github.com/9sheng/foobar-operator/cmd/foobar/options"
	actionclientset "github.com/9sheng/foobar-operator/pkg/client/action/clientset/versioned"
	actioninformers "github.com/9sheng/foobar-operator/pkg/client/action/informers/externalversions"
	"github.com/9sheng/foobar-operator/pkg/controller/foobar"
)

const (
	metricsEndpoint = "0.0.0.0:8080"
)

type FooBarServer struct {
	kubeconfig  string
	masterurl   string
	threadiness int
	Options     *options.Options
}

// Run the controllers
func (s *FooBarServer) Run(ctx context.Context) error {
	go http.ListenAndServe(metricsEndpoint, nil)

	cfg, err := clientcmd.BuildConfigFromFlags(s.masterurl, s.kubeconfig)
	if err != nil {
		klog.Fatalf("Build kubeconfig failed, %s", err.Error())
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Build kubeClient failed, %s", err.Error())
		return err
	}

	actionClient, err := actionclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Build actionClient failed, %s", err.Error())
		return err
	}

	serverVersion, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		klog.Fatalf("Failed to discover Kubernetes API server version: %v", err)
	} else {
		klog.Infof("Kubernetes API server version: %s", serverVersion)
	}

	period := metav1.Duration{
		Duration: 1 * time.Second,
	}
	// create all sharedInformerFactory
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, resyncPeriod(period)())
	actionInformerFactory := actioninformers.NewSharedInformerFactory(actionClient, resyncPeriod(period)())

	foobarController := foobar.NewController(
		kubeClient,
		actionClient,
		serverVersion,
		kubeInformerFactory.Core().V1().Pods(),
		actionInformerFactory.Test().V1().FooBars())

	go foobarController.Run(ctx, s.threadiness)

	<-ctx.Done()
	return nil
}

// resyncing with the api server.
func resyncPeriod(period metav1.Duration) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(period.Nanoseconds()) * factor)
	}
}
