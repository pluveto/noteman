package app

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

type BuildProcessor struct {
	appConf *AppConf
	cmd     *BuildCmd
}

func NewBuildProcessor(appConf *AppConf, cmd *BuildCmd) *BuildProcessor {
	return &BuildProcessor{
		appConf: appConf,
		cmd:     cmd,
	}
}

func (p *BuildProcessor) Execute() {
	logrus.Debugln("Build processor")

	cmd := exec.Command(p.appConf.Build.Command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = p.appConf.Build.WorkingDirectory
	logrus.Debugln("run command: ", cmd.String(), " in ", cmd.Dir)
	err := cmd.Run()
	if err != nil {
		logrus.Fatalf("failed to call cmd.Run(): %v", err)
	}

}
