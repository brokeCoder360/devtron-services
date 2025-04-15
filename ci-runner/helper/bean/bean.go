package bean

import "github.com/devtron-labs/common-lib/utils/bean"

type DockerBuildStageMetadata struct {
	TargetPlatforms []*bean.TargetPlatform `json:"targetPlatforms"`
}
