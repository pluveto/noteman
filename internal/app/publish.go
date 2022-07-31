package app

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pluveto/noteman/internal/pkg"
	"github.com/sirupsen/logrus"
)

type PublishProcessor struct {
	appConf *AppConf
	cmd     *PublishCmd
}

func NewPublishProcessor(appConf *AppConf, cmd *PublishCmd) *PublishProcessor {
	if _, err := os.Stat(appConf.Publish.Artifacts); err != nil {
		logrus.Fatalln("invalid artifacts dir: ", err.Error())
	}
	return &PublishProcessor{
		appConf: appConf,
		cmd:     cmd,
	}
}

func getTmpDir() string {
	tmpdir := path.Join(os.TempDir(), "noteman")
	err := os.MkdirAll(tmpdir, 0755)
	if err != nil {
		logrus.Fatalln("error when creating tmp dir: ", err)
	}
	return tmpdir
}

func (p *PublishProcessor) Execute() {
	arc := p.createTmpArch()
	defer func() {
		tmpdir := getTmpDir()
		os.RemoveAll(tmpdir)
	}()
	p.publish(arc)
}

func (p *PublishProcessor) publish(arc string) {
	svc := p.appConf.Publish.Service
	if svc.Name == "simple_http_upload" {
		headers := map[string]string{
			"authorization": svc.Params["auth"].(string),
		}
		resp, err := pkg.PostFile("file", arc, svc.Params["api"].(string), headers)
		if err != nil {
			logrus.Fatal("error when post file: ", err)
		}
		if resp.StatusCode != 200 {
			logrus.Errorln("unexpected response: ", resp.Status)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Errorln("failed to read resp body: ", err)
			} else {
				logrus.Infoln("resp: ", body)
			}
			os.Exit(1)
		}
	}
}
func (p *PublishProcessor) createTmpArch() string {
	tmpdir := getTmpDir()
	artifacts := p.appConf.Publish.Artifacts
	artifactId, err := pkg.FileMD5(path.Join(artifacts, "index.xml"))
	if err != nil {
		logrus.Fatalln("error when generating artifact id: ", err)
	}
	tmpArch := path.Join(tmpdir, artifactId+".zip")
	err = zipDir(artifacts, tmpArch)
	if err != nil {
		logrus.Fatalln("error when zip dir: ", err)
	}
	return tmpArch
}

func zipDir(baseFolder string, outPath string) (err error) {
	if !strings.HasSuffix(baseFolder, "/") && !strings.HasSuffix(baseFolder, string(os.PathSeparator)) {
		baseFolder += "/"
	}

	// Get a Buffer to Write To
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add some files to the archive.
	err = addFiles(w, baseFolder, "")
	if err != nil {
		return err
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func addFiles(w *zip.Writer, basePath, baseInZip string) error {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		// fmt.Println(basePath + file.Name())
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			// Add some files to the archive.
			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		} else if file.IsDir() {

			// Recurse
			newBase := basePath + file.Name() + "/"
			addFiles(w, newBase, baseInZip+file.Name()+"/")
		}
	}
	return nil
}
