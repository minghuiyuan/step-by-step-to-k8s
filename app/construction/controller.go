package construction

import (
	"fmt"
	clientset "k8s.io/client-go/kubernetes"
	"strings"
	"time"

	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	k8sInformers "k8s.io/client-go/informers"
	apiserverconfig "k8s.io/component-base/config/v1alpha1"
)

type KubePgServer struct {
	Kube2Cons
	Master            string
	Kubeconfig        string
}

type Controller struct {
	kubeClient         clientset.Interface
	k8sInformer        k8sInformers.SharedInformerFactory
	workerCount        int
}

func NewController(
	kubeClient clientset.Interface,
	k8sInformer k8sInformers.SharedInformerFactory,
	workerCount int,
) *Controller {
	return &Controller{
		kubeClient:         kubeClient,
		k8sInformer:        k8sInformer,
		workerCount:        workerCount,
	}
}

func (s *KubePgServer) AddFlags(fs *pflag.FlagSet, allControllers []string) {
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringArrayVar(&s.Controllers, "controllers", []string{"*"}, "the controllers to start up")
}

// Validate is used to validate the construction and config before launching the controller manager
func (s *KubePgServer) Validate(allControllers []string, disabledByDefaultControllers []string) error {
	var errs []error

	allControllersSet := sets.NewString(allControllers...)
	for _, controller := range s.Controllers {
		if controller == "*" {
			continue
		}
		if strings.HasPrefix(controller, "-") {
			controller = controller[1:]
		}

		if !allControllersSet.Has(controller) {
			errs = append(errs, fmt.Errorf("%q is not in the list of known controllers", controller))
		}
	}

	return utilerrors.NewAggregate(errs)
}
