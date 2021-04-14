package construction

import (
	"flag"
	"github.com/spf13/pflag"
	"time"

	k8sinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	kflag "k8s.io/component-base/cli/flag"
	klog "k8s.io/klog/v2"
)

type app struct {
	KubeConfig string
	workerCount int
}

// AddFlags adds flags
func (a *app) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&a.KubeConfig, "kubeconfig", a.KubeConfig, "Path to apiserver kubeconfig file with authorization and master location information.")
	fs.IntVar(&a.workerCount, "worker-count", a.workerCount, "number of workers")

	pflag.CommandLine.SetNormalizeFunc(kflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func (a *app) run() error {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", a.KubeConfig)
	if err != nil {
		klog.Fatalf("create kubeconfig failed: %v", err)
	}

	kubeClient, err := clientset.NewForConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Creating Kube Client failed: %v", err)
	}

	klog.V(3).Infof("kubeconfig:%s",kubeconfig)

	k8sInformerFactory := k8sinformers.NewSharedInformerFactory(kubeClient, 1*time.Minute)

	go NewController(kubeClient,k8sInformerFactory,1)

	return nil
}
