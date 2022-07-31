package main

import (
	"os"
	"path"

	"github.com/pluveto/noteman/internal/app"
	"github.com/pluveto/noteman/internal/pkg"

	"github.com/sirupsen/logrus"
)

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}
func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	path, err := pkg.NewPathSearcher().
		AddSearchPaths(cwd+"/examples", cwd, home, path.Join(home, ".config")).
		AddSearchNames("noteman.json", "noteman.jsonc", "noteman/config.json", "noteman/config.jsonc").
		Execute()

	if err != nil {
		panic("failed to get config: " + err.Error())
	}
	conf, err := app.AppConfFromFile(path)
	if err != nil {
		panic("failed to load config: " + err.Error() + " from" + path)
	}
	conf.SetConfPath(path)
	if conf.Target.Mapping == nil {
		conf.Target.Mapping = make(map[string]string)
	}
	conf.ResolveRelPath()
	logrus.Debugln("app conf: ", pkg.MustJsonEncode(conf))

	app.NewApp(conf).Run()
}
