package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
func (p *SyncProcessor) Execute() error {
	// Complete parsing, formatting and path validation before writing any files.
	metaouts := []*metaout{}
	targets := map[string]string{}
	bodies := map[string]string{}
	for _, sourcePath := range p.tasks {
		if !pkg.IsMarkdownExt(sourcePath) {
			continue
		}
		source, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("read %q: %w", sourcePath, err)
		}
		mb, err := pkg.ExtractMarkdownMeta([]rune(string(source)))
		if err != nil {
			return fmt.Errorf("extract metadata %q: %w", sourcePath, err)
		}
		mb.Meta = make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(mb.RawMeta), &mb.Meta); err != nil {
			return fmt.Errorf("parse metadata %q: %w", sourcePath, err)
		}
		if mb.Meta == nil {
			mb.Meta = make(map[string]interface{})
		}
		out := &metaout{srcPath: sourcePath, mb: mb}
		title, err := extractTitle(out)
		if err != nil {
			return fmt.Errorf("title %q: %w", sourcePath, err)
		}
		mb.Meta["title"] = title
		date, err := extractDate(out)
		if err != nil {
			return fmt.Errorf("date %q: %w", sourcePath, err)
		}
		mb.Meta["date"] = date
		slug, _, err := extractSlug(out)
		if err != nil {
			return fmt.Errorf("slug %q: %w", sourcePath, err)
		}
		mb.Meta["slug"] = slug
		if _, ok := mb.Meta["lang"]; !ok {
			mb.Meta["lang"] = "zh"
		}
		lang, ok := mb.Meta["lang"].(string)
		if !ok || lang == "" {
			return fmt.Errorf("lang must be a nonempty string in %q", sourcePath)
		}
		mathEnabled := false
		if value, exists := mb.Meta["mathjax"]; exists {
			mathEnabled, ok = value.(bool)
			if !ok {
				return fmt.Errorf("mathjax must be a boolean in %q", sourcePath)
			}
		}
		var buff bytes.Buffer
		if err := mdreformatter.Format([]byte(mb.RawBody), &buff, mathEnabled); err != nil {
			return fmt.Errorf("format %q: %w", sourcePath, err)
		}
		mb.RawBodyFormatted = buff.String()
		targetPath, err := p.appConf.Target.ResolveMapping(sourcePath, slug, lang)
		if err != nil {
			return fmt.Errorf("map %q: %w", sourcePath, err)
		}
		out.targetPath = targetPath + ".md"
		if previous, exists := targets[out.targetPath]; exists {
			return fmt.Errorf("duplicate target %q for %q and %q", out.targetPath, previous, sourcePath)
		}
		targets[out.targetPath] = sourcePath
		meta, err := yaml.Marshal(mb.Meta)
		if err != nil {
			return fmt.Errorf("encode metadata %q: %w", sourcePath, err)
		}
		if string(meta) != mb.RawMeta {
			mb.RawMeta = string(meta)
			mb.MetaChanged = true
		}
		hugoBody, err := dumpForHugo(mb)
		if err != nil {
			return fmt.Errorf("encode Hugo metadata %q: %w", sourcePath, err)
		}
		bodies[out.targetPath] = hugoBody
		metaouts = append(metaouts, out)
	}
	for _, out := range metaouts {
		if err := os.MkdirAll(filepath.Dir(out.targetPath), 0755); err != nil {
			return fmt.Errorf("create target directory %q: %w", out.targetPath, err)
		}
		if err := os.WriteFile(out.targetPath, []byte(bodies[out.targetPath]), 0644); err != nil {
			return fmt.Errorf("write target %q: %w", out.targetPath, err)
		}
	}
	if p.cmd.NoWriteBack {
		return nil
	}
	for _, out := range metaouts {
		if !out.mb.MetaChanged {
			continue
		}
		if err := os.WriteFile(out.srcPath, []byte(out.mb.Dump()), 0644); err != nil {
			return fmt.Errorf("write metadata %q: %w", out.srcPath, err)
		}
	}
	return nil
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

// dumpForHugo renders target markdown without reserved Hugo front matter keys.
// `lang` is kept on the source note for path mapping, but Hugo v0.144+ infers
// language from contentDir and errors if `lang` is set in front matter.
func dumpForHugo(mb *pkg.MarkdownMetaBody) (string, error) {
	meta := make(map[string]interface{}, len(mb.Meta))
	for k, v := range mb.Meta {
		if k == "lang" || k == "kind" || k == "path" {
			continue
		}
		meta[k] = v
	}
	b, err := yaml.Marshal(meta)
	if err != nil {
		return "", err
	}
	return "---\n" + string(b) + "---\n" + mb.RawBodyFormatted, nil
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
