package k8s_test

import (
	"io"
	"net/url"
	"os"
	"os/exec"
	"testing"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/test"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetes Backend Suite")
}

var logger = test.SetupLogging()

var testenv *envtest.Environment
var kcfg *os.File

var _ = BeforeSuite(func() {
	log.SetLogger(zapr.NewLogger(logger.Named("k8s")))

	// Setup envtest
	kubeBuilderAssets, err := exec.Command("setup-envtest", "use", "-p", "path").Output()
	Expect(err).To(Succeed(), "Failed to run setup-envtest. Please install it first:\n\n    go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest\n")

	testenv = &envtest.Environment{
		CRDDirectoryPaths:     []string{"../../../etc/kubernetes/crds"},
		BinaryAssetsDirectory: string(kubeBuilderAssets),
	}

	cfg, err := testenv.Start()
	Expect(err).To(Succeed())

	kcfg, err = os.CreateTemp("", "kubeconfig-*.yaml")
	Expect(err).To(Succeed())

	err = writeKubeconfig(cfg, kcfg)
	Expect(err).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())
})

var _ = Describe("Kubernetes backend", Label("broken-on-windows"), func() {
	var u url.URL

	BeforeEach(func() {
		u = url.URL{
			Scheme:   "k8s",
			Path:     kcfg.Name(),
			RawQuery: "namespace=default",
		}
	})

	test.BackendTest(&u, 2)
})

func writeKubeconfig(rc *rest.Config, wr io.Writer) error {
	ac := api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			"default-cluster": {
				Server:                   rc.Host,
				CertificateAuthorityData: rc.TLSClientConfig.CAData,
			},
		},
		Contexts: map[string]*api.Context{
			"default-context": {
				AuthInfo:  "default-auth",
				Cluster:   "default-cluster",
				Namespace: "default",
			},
		},
		CurrentContext: "default-context",
		AuthInfos: map[string]*api.AuthInfo{
			"default-auth": {
				ClientCertificateData: rc.TLSClientConfig.CertData,
				ClientKeyData:         rc.TLSClientConfig.KeyData,
			},
		},
	}

	acfg, err := clientcmd.Write(ac)
	if err != nil {
		return err
	}

	if _, err := wr.Write(acfg); err != nil {
		return err
	}

	return nil
}
