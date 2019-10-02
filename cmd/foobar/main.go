package main

import (
	"context"
	"flag"
	"os"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/9sheng/foobar-operator/cmd/foobar/app"
	"github.com/9sheng/foobar-operator/cmd/foobar/options"
	"github.com/9sheng/foobar-operator/pkg/signals"
	"github.com/9sheng/foobar-operator/pkg/version"
)

const (
	metricsEndpoint = "0.0.0.0:8080"
)

var (
	kubeConfig string
	masterURL  string
	appConfig  string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Infof("GitVersion: %s, BuildTime: %s", version.GetBuildVersion(), version.GetBuildTime())
	klog.Infof("kubeconfig=%s", kubeConfig)
	klog.Infof("appconfig=%s", appConfig)
	klog.Infof("Starting foobar-operator server")

	ctx, cancelFunc := context.WithCancel(context.Background())
	signals.SetupSignalHandler(cancelFunc)

	options, err := options.Load(appConfig)
	if err != nil {
		klog.Fatalf("LoadDomainServer %s failed: %s", appConfig, err.Error())
		klog.Flush()
		os.Exit(1)
	}

	server := &app.FooBarServer{
		Options: options,
	}

	if err := server.Run(ctx); err != nil {
		klog.Fatalf("run foobar-operator failed: %s", err.Error())
		klog.Flush()
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&appConfig, "appconfig", "", "Path to app config.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
