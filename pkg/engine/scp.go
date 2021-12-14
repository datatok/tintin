package engine

import (
	"fmt"

	"github.com/datatok/tintin/pkg/executions"
	"github.com/datatok/tintin/pkg/utils/constant"
)

func SCPBuildDisplay(stage executions.StageHit) string {
	phase := stage.PostCheck

	if stage.PostCheck.Status == constant.DoneOk {
		if len(phase.Meta.Display) == 0 {
			return fmt.Sprintf("File of %s", phase.Meta.Size)
		}
	}

	if len(phase.Meta.Display) == 0 && len(phase.Meta.Reason) > 0 {
		return phase.Meta.Reason
	}

	return phase.Meta.Display
}
