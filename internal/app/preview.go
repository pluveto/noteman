package app

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

type PreviewProcessor struct {
	appConf *AppConf
	cmd     *PreviewCmd
}

func NewPreviewProcessor(appConf *AppConf, cmd *PreviewCmd) *PreviewProcessor {
	return &PreviewProcessor{
		appConf: appConf,
		cmd:     cmd,
	}
}

func (p *PreviewProcessor) Execute() {
	logrus.Debugln("Preview processor")

	cmd := exec.Command(p.appConf.Preview.Command, p.appConf.Preview.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = p.appConf.Preview.WorkingDirectory
	logrus.Debugln("run preview command: ", cmd.String(), " in ", cmd.Dir)
	err := cmd.Run()
	if err != nil {
		logrus.Fatalf("failed to call cmd.Run(): %v", err)
	}

}
