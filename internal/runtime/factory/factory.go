package factory

import (
	"fmt"

	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
	"github.com/HXong/predeploy-guard/internal/runtime/compose"
)

func NewAdapter(runtimeType string) (predeployruntime.Adapter, error) {
	switch predeployruntime.Type(runtimeType) {
	case "", predeployruntime.TypeDockerCompose:
		return compose.New(), nil
	default:
		return nil, fmt.Errorf(
			"unsupported runtime.type %q; supported runtimes: %s",
			runtimeType,
			predeployruntime.TypeDockerCompose,
		)
	}
}
