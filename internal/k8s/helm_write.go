package k8s

import (
	"context"
	"fmt"

	"github.com/gnana997/periscope/internal/clusters"
	"github.com/gnana997/periscope/internal/credentials"

	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type simpleRESTClientGetter struct {
	cfg       *rest.Config
	namespace string
}

func (s *simpleRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return s.cfg, nil
}

func (s *simpleRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	c, err := discovery.NewDiscoveryClientForConfig(s.cfg)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(c), nil
}

func (s *simpleRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	c, err := s.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(c)
	expander := restmapper.NewShortcutExpander(mapper, c, nil)
	return expander, nil
}

func (s *simpleRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(*clientcmdapi.NewConfig(), &clientcmd.ConfigOverrides{})
}

// RollbackHelmRelease executes a helm rollback to the specified revision using the official Helm SDK.
func RollbackHelmRelease(ctx context.Context, p credentials.Provider, c clusters.Cluster, namespace, name string, revision int) error {
	cfg, err := buildRestConfig(ctx, p, c)
	if err != nil {
		return fmt.Errorf("build rest config: %w", err)
	}

	cs, err := newClientFn(ctx, p, c)
	if err != nil {
		return fmt.Errorf("build clientset: %w", err)
	}

	// Re-use our existing probe to determine if releases are stored in Secrets or ConfigMaps
	drv, err := resolveHelmDriver(ctx, cs, c)
	if err != nil {
		return fmt.Errorf("resolve helm driver: %w", err)
	}

	actionConfig := new(action.Configuration)
	getter := &simpleRESTClientGetter{cfg: cfg, namespace: namespace}
	
	if err := actionConfig.Init(getter, namespace, drv, func(format string, v ...interface{}) {}); err != nil {
		return fmt.Errorf("init helm action config: %w", err)
	}

	client := action.NewRollback(actionConfig)
	client.Version = revision

	if err := client.Run(name); err != nil {
		return fmt.Errorf("helm rollback: %w", err)
	}

	return nil
}
