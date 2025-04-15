package helper

import util2 "github.com/devtron-labs/ci-runner/executor/util"

func SetKeyValueInGlobalSystemEnv(scriptEnvs *util2.ScriptEnvVariables, key, value string) {
	scriptEnvs.SystemEnv[key] = value
}
