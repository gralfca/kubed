package e2e_test

import (
	"os"
	"testing"
	"time"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/kubed/test/e2e/framework"
	searchlightcheme "github.com/appscode/searchlight/client/clientset/versioned/scheme"
	voyagerscheme "github.com/appscode/voyager/client/clientset/versioned/scheme"
	kubedbscheme "github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	rbac "k8s.io/api/rbac/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/tools/clientcmd"
	prom_util "kmodules.xyz/monitoring-agent-api/prometheus/v1"
	stashscheme "stash.appscode.dev/stash/client/clientset/versioned/scheme"
)

const TestTimeout = 3 * time.Minute

var (
	root            *framework.Framework
	userRule        *rbac.ClusterRole
	userRoleBinding *rbac.ClusterRoleBinding
)

func TestE2E(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TestTimeout)
	junitReporter := reporters.NewJUnitReporter("report.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Kubed E2E Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	voyagerscheme.AddToScheme(clientsetscheme.Scheme)
	searchlightcheme.AddToScheme(clientsetscheme.Scheme)
	stashscheme.AddToScheme(clientsetscheme.Scheme)
	kubedbscheme.AddToScheme(clientsetscheme.Scheme)
	prom_util.AddToScheme(clientsetscheme.Scheme)

	clientConfig, err := clientcmd.BuildConfigFromContext(options.KubeConfig, options.KubeContext)
	Expect(err).NotTo(HaveOccurred())

	root = framework.New(clientConfig)
	root.KubeConfigPath = options.KubeConfig
	root.SelfHostedOperator = options.SelfHostedOperator

	By("Using Namespace " + root.Namespace())
	err = root.EnsureNamespace()
	Expect(err).NotTo(HaveOccurred())

	By("Creating CRDs")
	err = root.EnsureCreatedCRDs()
	Expect(err).NotTo(HaveOccurred())

	By("Creating initial kubed configuration file")
	err = framework.APIServerClusterConfig().Save(framework.KubedTestConfigFileDir)
	Expect(err).NotTo(HaveOccurred())

	if !root.SelfHostedOperator {
		By("Registering API service")
		err = root.Invoke().RegisterAPIService()
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	if !root.SelfHostedOperator {
		By("Cleaning API service stuff")
		root.Invoke().DeleteAPIService()
	}
	root.DeleteNamespace(root.Namespace())
	os.Remove(framework.KubedTestConfigFileDir)
})
