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

type Kube2Cons struct {
	v1.TypeMeta
	StorageProviders []string
	PgDataBase       string
	PgUser           string
	PgPassword       string
	EnableProfiling  bool
	PgHost           string
	PgPort           int
	EsHost           []string
	EsUser           string
	EsPassword       string
	DbConfigFile     string
	Address          string
	Port             int
	// contentType is contentType of requests sent to apiserver.
	ContentType string
	// kubeAPIQPS is the QPS to use while talking with kubernetes apiserver.
	KubeAPIQPS float32
	// kubeAPIBurst is the burst to use while talking with kubernetes apiserver.
	KubeAPIBurst int32
	// minResyncPeriod is the resync period in reflectors; will be random between
	// minResyncPeriod and 2*minResyncPeriod.
	MinResyncPeriod v1.Duration
	// serviceAccountKeyFile is the filename containing a PEM-encoded private RSA key
	// used to sign service account tokens.
	ServiceAccountKeyFile string
	// useServiceAccountCredentials indicates whether controllers should be run with
	// individual service account credentials.
	UseServiceAccountCredentials bool
	// leaderElection defines the configuration of leader election client.
	LeaderElection apiserverconfig.LeaderElectionConfiguration
	// rootCAFile is the root certificate authority will be included in service
	// account's token secret. This must be a valid PEM-encoded CA bundle.
	RootCAFile string
	// concurrentSATokenSyncs is the number of service account token syncing operations
	// that will be done concurrently.
	ConcurrentSATokenSyncs int32
	// How long to wait between starting controller managers
	ControllerStartInterval v1.Duration
	Controllers             []string
	//API GVK blacklist. A comma separated API group version kind
	APIGVKBlackList []string

	//API sub resources blacklist. A comma separated API sub resources list
	APIBlackListSubResources []string

	PurgeFrequency time.Duration
	PurgeLimit     time.Duration
	SlackChannel   string
	ProxyAddr      string
	EnableProxy    bool
	TableList      []string

	FullSyncFrequency     time.Duration
	BulkPostgresThreshold int
	ApiServerType         string
	EnableMainControllers bool
	EnableCrdControllers  bool
	ForHistory            bool
	Usage                 string
}

type KubePgServer struct {
	Kube2Cons
	Master            string
	Kubeconfig        string
	TMKubeconfig      string
	TessnetKubeconfig string
	FedKubeconfig     string
	CmsClientConfig   string
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
