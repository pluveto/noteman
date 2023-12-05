package app

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/djherbis/times"
	slugUtil "github.com/gosimple/slug"

	"github.com/pluveto/noteman/internal/pkg"
	"github.com/pluveto/noteman/internal/pkg/mdreformatter"
	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v2"
)

type SyncProcessor struct {
	appConf *AppConf
	cmd     *SyncCmd
	tasks   []string
}

type SyncHandler = func() error
type SyncPipeline struct {
	Handlers []SyncHandler
}

func NewSyncPipeline() *SyncPipeline {
	return &SyncPipeline{}
}

func (p *SyncPipeline) Execute() {
	logrus.Debugln("Sync pipeline")
}

// NewSyncProcessor creates a new SyncProcessor.
func NewSyncProcessor(appConf *AppConf, cmd *SyncCmd) *SyncProcessor {
	ret := &SyncProcessor{
		appConf: appConf,
		cmd:     cmd,
		tasks:   []string{},
	}
	err := ret.PrepareTasks()
	if err != nil {
		logrus.Fatalln("failed at preparing tasks: ", err)
		os.Exit(1)
	}
	return ret
}
func (p *SyncProcessor) PrepareTasks() error {
	confDir := filepath.Dir(p.appConf.GetConfPath())
	sg := NewSourceGlobber(&p.appConf.Source, confDir)
	files, err := sg.Glob()
	if err != nil {
		return err
	}
	logrus.Debugln("files to be tasks")
	for i, file := range files {
		logrus.Debugln(i, file)
	}
	p.tasks = files
	return nil
}

type metaout struct {
	srcPath    string
	targetPath string
	mb         *pkg.MarkdownMetaBody
}

// Execute executes the SyncProcessor.
func (p *SyncProcessor) Execute() {
	logrus.Debugln("Sync processor")
	metaouts := []*metaout{}
	// pass 1 - load posts and meta
	for _, p := range p.tasks {
		source, err := ioutil.ReadFile(p)
		if !pkg.IsMarkdownExt(p) {
			logrus.Warningln("skipping non-markdown file: ", p)
			continue
		}
		if err != nil {
			logrus.Errorln("failed to read", pkg.QuotePath(p), err.Error())
			pkg.WaitForEnter()
			continue
		}
		splited, err := pkg.ExtractMarkdownMeta([]rune(string(source)))
		if err != nil {
			logrus.Errorln("failed to extract meta", pkg.QuotePath(p), err.Error())
			pkg.WaitForEnter()
			continue
		}
		metaouts = append(metaouts, &metaout{srcPath: p, mb: splited})
		logrus.Debugln("meta: ", splited.RawMeta)
		splited.Meta = make(map[string]interface{})
		// logrus.Debugln("body: ", splited.Body)
	}
	// pass 2 - fill meta
	/**
	 * required fields: title, date, slug
	 */
	for _, out := range metaouts {
		mb := out.mb
		pkg.Assert(mb.Meta != nil, "meta is nil")
		err := yaml.Unmarshal([]byte(mb.RawMeta), &mb.Meta)
		if err != nil {
			logrus.Errorln("failed to unmarshal meta", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		title, err := extractTitle(out)
		if err != nil {
			logrus.Errorln("failed to generate title", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		mb.Meta["title"] = title
		date, err := extractDate(out)
		if err != nil {
			logrus.Errorln("failed to get document date", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		mb.Meta["date"] = date
		slug, _, err := extractSlug(out)
		if err != nil {
			logrus.Errorln("failed to generate slug", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		mb.Meta["slug"] = slug
		// if generated {
		// 	// if slug is generated, write back to original file
		// 	logrus.Debugln("writing back to original file")
		// 	err := ioutil.WriteFile(out.srcPath, []byte(out.mb.Dump()), 0644)
		// 	if err != nil {
		// 		logrus.Errorln("failed to write meta back to source:", pkg.QuotePath(out.srcPath), err.Error())
		// 	}
		// }
	}
	// pass 3 - remove extra title and rerender body
	for _, out := range metaouts {
		raw := out.mb.RawBody
		var buff bytes.Buffer
		mathjaxEnabled := out.mb.Meta["mathjax"] != nil && out.mb.Meta["mathjax"].(bool)
		err := mdreformatter.Format([]byte(raw), &buff, mathjaxEnabled)
		if err != nil {
			logrus.Errorln("failed to reformat body", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		out.mb.RawBodyFormatted = buff.String()
	}
	// pass 4 - generate target path
	for i, out := range metaouts {
		slug_ := out.mb.Meta["slug"].(string)
		targetPath, err := p.appConf.Target.ResolveMapping(out.srcPath, slug_)
		if err != nil {
			logrus.Errorln("failed to resolve mapping", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		pkg.Assert(targetPath != "", "target path should not be empty")
		metaouts[i].targetPath = targetPath + ".md"
	}
	// pass 5 - write to target
	for _, out := range metaouts {
		mb := out.mb
		meta, err := yaml.Marshal(mb.Meta)
		if err != nil {
			logrus.Errorln("failed to marshal meta", pkg.QuotePath(out.srcPath), err.Error())
			pkg.WaitForEnter()
			continue
		}
		metaStr := string(meta)
		if strings.Compare(metaStr, mb.RawMeta) != 0 {
			mb.RawMeta = metaStr
			mb.MetaChanged = true
		}
		logrus.Debugln("target: ", out.targetPath)
		err = ioutil.WriteFile(out.targetPath, []byte(out.mb.DumpFormatted()), 0644)
		if err != nil {
			logrus.Errorln("failed to write meta, src:", pkg.QuotePath(out.srcPath), "target:", out.targetPath, err.Error())
			pkg.WaitForEnter()
			continue
		}
		logrus.Debugln("wrote meta from src:", pkg.QuotePath(out.srcPath), "target:", out.targetPath)
	}
	// pass 6 - write meta back to source
	for _, out := range metaouts {
		if !out.mb.MetaChanged {
			logrus.Debugln("no meta changed, skip write back to ", pkg.QuotePath(out.srcPath))
			continue
		}
		err := ioutil.WriteFile(out.srcPath, []byte(out.mb.Dump()), 0644)
		if err != nil {
			logrus.Errorln("failed to write meta back to source:", pkg.QuotePath(out.srcPath), err.Error())
		}
		logrus.Debugln("wrote meta back to source:", pkg.QuotePath(out.srcPath))
	}
}

// 首先尝试 title 字段，如果没有，则使用一级标题
func extractTitle(out *metaout) (title string, err error) {
	mb := out.mb
	if t, ok := mb.Meta["title"]; ok {
		if ts, ok := t.(string); ok {
			title = ts
		} else {
			err = errors.New("title is not string ")
			return
		}
	}
	if title == "" {
		bytes_ := []byte(mb.RawBody)
		reader := text.NewReader(bytes_)
		mdAst := goldmark.DefaultParser().Parse(reader)
		headAst := pkg.MdFindFirstHeading(mdAst)
		if headAst != nil {
			title = string((*headAst).(*ast.Heading).Text(bytes_))
		}
	}
	if title == "" {
		title = pkg.BaseNoExt(out.srcPath)
	}
	if title == "" {
		err = errors.New("title is empty ")
		return
	}
	logrus.Debugln("title: ", title)
	return title, nil
}

func extractDate(out *metaout) (string, error) {
	mb := out.mb
	if d, ok := mb.Meta["date"]; ok {
		if ts, ok := d.(string); ok {
			return ts, nil
		}
	}
	// use file created time
	t, err := times.Stat(out.srcPath)
	if err != nil {
		return "", err
	}
	if t.HasBirthTime() {
		return t.BirthTime().Format(time.RFC3339Nano), nil
	}
	if t.HasChangeTime() {
		return t.ChangeTime().Format(time.RFC3339Nano), nil
	}
	return "", errors.New("no date found")
}

// extractSlug 从 meta header 提取 slug，如果没有，则自动翻译一个
func extractSlug(out *metaout) (slug string, generated bool, err error) {
	mb := out.mb
	if s, ok := mb.Meta["slug"]; ok {
		if ts, ok := s.(string); ok {
			return ts, false, nil
		}
	}
	var slug_ string
	title := out.mb.Meta["title"].(string)
	if !pkg.IsPureASCII(title) {
		fmt.Println("Warning: Title may not in English, and slug is not provided.")
		fmt.Println("Do you want to provide a slug for this post?")
		sug_slug := suggestSlug(title)
		fmt.Print("Your slug (empty to use suggestion): ")
		fmt.Scanln(&slug_)
		if slug_ == "" {
			slug_ = sug_slug
		}
		if slug_ == "" {
			slug_ = slugUtil.Make(title)
		}
	} else {
		slug_ = slugUtil.Make(title)
	}
	pkg.Assert(slug_ != "", "slug is empty")
	logrus.Debugln("slug: ", slug_)
	return slug_, true, nil
}

func suggestSlug(title string) string {
	en, err := pkg.Translate(title, "zh", "en")
	if err != nil {
		logrus.Errorln("failed to translate title to english: ", err.Error())
		return ""
	}
	slug_ := slugUtil.Make(en)
	fmt.Printf("Suggested slug: ")
	fmt.Println(slug_)
	return slug_
}
