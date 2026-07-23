package factory

import (
	"fmt"

	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
	"github.com/HXong/predeploy-guard/internal/runtime/compose"
	"github.com/HXong/predeploy-guard/internal/runtime/kubernetes"
)

func NewAdapter(runtimeType string) (predeployruntime.Adapter, error) {
	switch predeployruntime.Type(runtimeType) {
	case "", predeployruntime.TypeDockerCompose:
		return compose.New(), nil
	case predeployruntime.TypeKubernetes:
		return kubernetes.New(), nil
	default:
		return nil, fmt.Errorf(
			"unsupported runtime.type %q; supported runtimes: %s, %s",
			runtimeType,
			predeployruntime.TypeDockerCompose,
			predeployruntime.TypeKubernetes,
		)
	}
}
