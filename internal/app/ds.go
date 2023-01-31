package app

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pluveto/noteman/internal/pkg"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/sirupsen/logrus"
)

type AppConfSource struct {
	Directories []string `json:"directories"`
	Filters     struct {
	} `json:"filters"`
}
type AppConfTarget struct {
	Mapping map[string]string `json:"mapping"`
}

func (t *AppConfTarget) ResolveMapping(sourcePath string, slug_ string) (targetPath string, err error) {
	pkg.Assert(t.Mapping != nil, "nil t.Mapping")
	sourcePath = filepath.Dir(sourcePath)
	// find most match prefix
	var prefix string
	for mapKey := range t.Mapping {
		// uniform path
		k := filepath.ToSlash(mapKey)
		sourcePath = filepath.ToSlash(sourcePath)
		if strings.HasPrefix(sourcePath, k) && len(k) > len(prefix) {
			prefix = k
		}
	}
	if prefix == "" {
		return "", errors.New("no match mapping prefix found for " + sourcePath)
	}
	target, ok := t.Mapping[prefix]
	if !ok {
		return "", errors.New("no mapping found for " + sourcePath)
	}
	return filepath.Join(target, slug_), nil
}

type AppConfBuild struct {
	Command          string   `json:"command"`
	Args             []string `json:"args"`
	WorkingDirectory string   `json:"working_directory"`
}
type AppConfPreview struct {
	Command          string   `json:"command"`
	Args             []string `json:"args"`
	WorkingDirectory string   `json:"working_directory"`
}
type PublishServiceConf struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

type AppConfPublish struct {
	Artifacts  string             `json:"artifacts"`
	Service    PublishServiceConf `json:"service"`
	PreviewUrl string             `json:"preview_url"`
}

type AppConf struct {
	confPath string         `json:"-"` // used for tracking path of current config file
	Source   AppConfSource  `json:"source"`
	Target   AppConfTarget  `json:"target"`
	Build    AppConfBuild   `json:"build"`
	Preview  AppConfPreview `json:"preview"`
	Publish  AppConfPublish
}

func (c *AppConf) SetConfPath(path string) {
	c.confPath = path
}

func (c *AppConf) GetConfPath() string {
	return c.confPath
}

func (c *AppConf) ResolveRelPath() {
	confDir := filepath.Dir(c.GetConfPath())
	pkg.Assert(confDir != "", "confDir is empty")

	for i, dir := range c.Source.Directories {
		dir = pkg.ReplaceDotPath(dir, confDir) // convert to absolute path
		c.Source.Directories[i] = dir
	}
	pkg.Assert(c.Target.Mapping != nil, "nil c.Target.Mapping")
	for k, v := range c.Target.Mapping {
		v = pkg.ReplaceDotPath(v, confDir) // convert to absolute path
		newKey := pkg.ReplaceDotPath(k, confDir)
		if newKey != k {
			delete(c.Target.Mapping, k)
			c.Target.Mapping[newKey] = v
		} else {
			c.Target.Mapping[k] = v
		}
	}
}

type SyncCmd struct {
}

type BuildCmd struct{}

type PublishCmd struct{}
type PreviewCmd struct{}

type CliArgs struct {
	SyncCmd    *SyncCmd    `arg:"subcommand:sync"`
	BuildCmd   *BuildCmd   `arg:"subcommand:build"`
	PublishCmd *PublishCmd `arg:"subcommand:publish"`
	PreviewCmd *PreviewCmd `arg:"subcommand:preview"`
}

type SourceGlobber struct {
	Source     *AppConfSource
	workingDir string
}

// NewSourceGlobber creates a new SourceGlobber.
func NewSourceGlobber(source *AppConfSource, workingDir string) *SourceGlobber {
	return &SourceGlobber{
		Source:     source,
		workingDir: workingDir,
	}
}

// Glob returns a list of files.
func (g *SourceGlobber) Glob() ([]string, error) {
	ret := []string{}
	for _, dir := range g.Source.Directories {
		out := []string{}
		pkg.SimpleGlob(dir, &out)

		// if has .gitignore file, apply its rules
		if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil {
			rules, err := ignore.CompileIgnoreFile(filepath.Join(dir, ".gitignore"))
			if err != nil {
				return nil, errors.New("glob failed to compile .gitignore: " + err.Error())
			}
			for _, path := range out {
				// remove g.workingDir prefix from path
				abspath := path
				path = strings.TrimPrefix(path, g.workingDir)
				if strings.HasSuffix(path, ".gitignore") {
					continue
				}
				// if ignore
				if rules.MatchesPath(path) {
					logrus.Debugln("ignoring: ", path, "by .gitignore rule")
					continue
				}
				// if not match, append it
				ret = append(ret, abspath)
			}
		} else {
			ret = append(ret, out...)
		}
	}
	sort.Strings(ret)
	pkg.RemoveDupSortedSorted(&ret)

	return ret, nil
}
