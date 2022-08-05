package app

import (
	"encoding/json"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/pluveto/noteman/internal/pkg"
	"github.com/sirupsen/logrus"
)

func AppConfFromFile(path string) (*AppConf, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	jsoncRaw := string(dat)
	var conf AppConf
	jsonRaw := pkg.RemoveComment(jsoncRaw)
	err = json.Unmarshal([]byte(jsonRaw), &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

type App struct {
	Conf    *AppConf
	CliArgs *CliArgs
}

func NewApp(conf *AppConf) *App {
	var args CliArgs
	return &App{
		Conf:    conf,
		CliArgs: &args,
	}
}

func (a *App) Run() {
	logrus.Info("noteman is started")
	p := arg.MustParse(a.CliArgs)
	if p.Subcommand() == nil {
		p.Fail("missing subcommand")
	}
	if a.CliArgs.SyncCmd != nil {
		NewSyncProcessor(a.Conf, a.CliArgs.SyncCmd).Execute()
	}
	if a.CliArgs.BuildCmd != nil {
		NewBuildProcessor(a.Conf, a.CliArgs.BuildCmd).Execute()
	}
	if a.CliArgs.PublishCmd != nil {
		NewPublishProcessor(a.Conf, a.CliArgs.PublishCmd).Execute()
	}
	if a.CliArgs.PreviewCmd != nil {
		NewPreviewProcessor(a.Conf, a.CliArgs.PreviewCmd).Execute()
	}
}
