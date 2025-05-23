package template

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/telecom-cloud/crafter/pkg/meta"
	"github.com/telecom-cloud/crafter/pkg/util"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
)

var DefaultDelimiters = [2]string{"{{", "}}"}

type Config struct {
	Layouts []Template `yaml:"layouts"`
}

const (
	Skip   = "skip"
	Cover  = "cover"
	Append = "append"
)

type Template struct {
	Default        bool           // Is it the default template
	Path           string         `yaml:"path"`            // The generated path and its filename.
	Delims         [2]string      `yaml:"delims"`          // Template Action Instruction Identifier, default: "{{}}"
	Body           string         `yaml:"body"`            // Render template, currently only supports go template syntax
	Disable        bool           `yaml:"disable"`         // Disable generating file, used to disable default package template
	LoopMethod     bool           `yaml:"loop_method"`     // Loop generate files based on "method"
	LoopService    bool           `yaml:"loop_service"`    // Loop generate files based on "service"
	UpdateBehavior UpdateBehavior `yaml:"update_behavior"` // Update command behavior; 0:unchanged, 1:regenerate, 2:append
}

type UpdateBehavior struct {
	Type string `yaml:"type"` // Update behavior type: skip/cover/append
	// the following variables are used for append update
	AppendKey      string   `yaml:"append_key"`         // Append content based in key; for example: 'method'/'service'
	InsertKey      string   `yaml:"insert_key"`         // Insert content by "insert_key"
	AppendTpl      string   `yaml:"append_content_tpl"` // Append content if UpdateBehavior is "append"
	ImportTpl      []string `yaml:"import_tpl"`         // Import insert template
	AppendLocation string   `yaml:"append_location"`    // AppendLocation specifies the location of append,  the default is the end of the file
}

// TemplateGenerator contains information about the output template
type TemplateGenerator struct {
	OutputDir     string
	Config        *Config
	IsPackageTpl  bool
	Excludes      []string
	Templates     map[string]*TemplateInfo
	Dirs          map[string]bool
	files         []util.File
	excludedFiles map[string]*util.File
}

type TemplateInfo struct {
	Template *template.Template
	Info     Template
}

func (tg *TemplateGenerator) Init() error {
	if tg.Config == nil {
		return errors.New("config not set yet")
	}

	if tg.Templates == nil {
		tg.Templates = make(map[string]*TemplateInfo, len(tg.Config.Layouts))
	}
	if tg.Dirs == nil {
		tg.Dirs = make(map[string]bool)
	}

	for _, l := range tg.Config.Layouts {
		if tg.IsPackageTpl && IsDefaultPackageTpl(l.Path) {
			continue
		}

		// check if is a directory
		var noFile bool
		if strings.HasSuffix(l.Path, string(filepath.Separator)) {
			noFile = true
		}
		path := l.Path
		if filepath.IsAbs(path) {
			return fmt.Errorf("absolute template path '%s' is not allowed", path)
		}
		dir := filepath.Dir(path)
		isExist, err := util.PathExist(filepath.Join(tg.OutputDir, dir))
		if err != nil {
			return fmt.Errorf("check directory '%s' failed, err: %v", dir, err.Error())
		}
		tg.Dirs[dir] = false
		if isExist {
			tg.Dirs[dir] = true
		}

		if noFile {
			continue
		}

		// parse templates
		if _, ok := tg.Templates[path]; ok {
			continue
		}
		err = tg.LoadLayout(l, path, false)
		if err != nil {
			return err
		}
	}

	excludes := make(map[string]*util.File, len(tg.Excludes))
	for _, f := range tg.Excludes {
		excludes[f] = &util.File{}
	}

	tg.excludedFiles = excludes
	return nil
}

func (tg *TemplateGenerator) LoadLayout(layout Template, tplName string, isDefaultTpl bool) error {
	delims := DefaultDelimiters
	// set template delims
	if layout.Delims[0] != "" && layout.Delims[1] != "" {
		delims = layout.Delims
	}
	// insert template funcs and delims
	tpl := template.New(tplName).Funcs(FuncMap).Delims(delims[0], delims[1])
	var err error
	if tpl, err = tpl.Parse(layout.Body); err != nil {
		return fmt.Errorf("parse template '%s' failed, err: %v", tplName, err.Error())
	}
	layout.Default = isDefaultTpl
	tg.Templates[tplName] = &TemplateInfo{
		Template: tpl,
		Info:     layout,
	}
	return nil
}

func (tg *TemplateGenerator) Generate(input interface{}, tplName, filepath string, noRepeat bool) error {
	if tplName != "" {
		return tg.generateWithSpecifiedTplName(input, tplName, filepath, noRepeat)
	}

	return tg.generateByTraverseTpl(input, noRepeat)
}

func (tg *TemplateGenerator) generateWithSpecifiedTplName(input interface{}, tplName, filepath string, noRepeat bool) error {
	file := bytes.NewBuffer(nil)
	info := tg.Templates[tplName]
	if info == nil {
		return fmt.Errorf("tpl %s not found, %v", tplName, tg.Templates)
	}
	tpl := info.Template
	if tpl == nil {
		return fmt.Errorf("tpl %s not found", tplName)
	}
	if err := tpl.Execute(file, input); err != nil {
		return fmt.Errorf("render template '%s' failed, err: %v", tplName, err.Error())
	}

	in := util.File{filepath, string(file.Bytes()), noRepeat, tplName}
	tg.files = append(tg.files, in)
	return nil
}

func (tg *TemplateGenerator) generateByTraverseTpl(input interface{}, noRepeat bool) error {
	// check if "*" (global scope) data exists, and stores it to all
	var all map[string]interface{}
	if data, ok := input.(map[string]interface{}); ok {
		ad, ok := data["*"]
		if ok {
			all = ad.(map[string]interface{})
		}
		if all == nil {
			all = map[string]interface{}{}
		}
		all["cftVersion"] = meta.Version
	}

	file := bytes.NewBuffer(nil)
	for path, tpl := range tg.Templates {
		file.Reset()
		var fd interface{}
		// search and merge rendering data
		if data, ok := input.(map[string]interface{}); ok {
			td := map[string]interface{}{}
			tmp, ok := data[path]
			if ok {
				td = tmp.(map[string]interface{})
			}
			for k, v := range all {
				td[k] = v
			}
			fd = td
		} else {
			fd = input
		}
		if err := tpl.Template.Execute(file, fd); err != nil {
			return fmt.Errorf("render template '%s' failed, err: %v", path, err.Error())
		}

		in := util.File{Path: path, Content: string(file.Bytes()), NoRepeat: noRepeat, FileTplName: tpl.Template.Name()}
		tg.files = append(tg.files, in)
	}
	return nil
}

func (tg *TemplateGenerator) Persist() error {
	files := tg.files
	outPath := tg.OutputDir
	if !filepath.IsAbs(outPath) {
		outPath, _ = filepath.Abs(outPath)
	}

	for _, data := range files {
		// check for -E flags
		if _, ok := tg.excludedFiles[filepath.Join(data.Path)]; ok {
			continue
		}

		// lint file
		if err := data.Lint(); err != nil {
			return err
		}

		// create rendered file
		abPath := filepath.Join(outPath, data.Path)
		abDir := filepath.Dir(abPath)
		isExist, err := util.PathExist(abDir)
		if err != nil {
			return fmt.Errorf("check directory '%s' failed, err: %v", abDir, err.Error())
		}
		if !isExist {
			if err := os.MkdirAll(abDir, os.FileMode(0o744)); err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", abDir, err.Error())
			}
		}

		err = func() error {
			file, err := os.OpenFile(abPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(0o755))
			defer file.Close()
			if err != nil {
				return fmt.Errorf("open file '%s' failed, err: %v", abPath, err.Error())
			}
			if _, err = file.WriteString(data.Content); err != nil {
				return fmt.Errorf("write file '%s' failed, err: %v", abPath, err.Error())
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	tg.files = tg.files[:0]
	return nil
}

func (tg *TemplateGenerator) GetFormatAndExcludedFiles() ([]util.File, error) {
	var files []util.File
	outPath := tg.OutputDir
	if !filepath.IsAbs(outPath) {
		outPath, _ = filepath.Abs(outPath)
	}

	for _, data := range tg.Files() {
		if _, ok := tg.excludedFiles[filepath.Join(data.Path)]; ok {
			continue
		}

		// check repeat files
		logs.Infof("Write %s", data.Path)
		isExist, err := util.PathExist(filepath.Join(data.Path))
		if err != nil {
			return nil, fmt.Errorf("check file '%s' failed, err: %v", data.Path, err.Error())
		}
		if isExist && data.NoRepeat {
			logs.Warnf("file '%s' already exists, so drop the generated file", data.Path)
			continue
		}

		// lint file
		if err := data.Lint(); err != nil {
			logs.Warnf("Lint file: %s failed:\n %s\n", data.Path, data.Content)
		}
		files = append(files, data)
	}

	return files, nil
}

func (tg *TemplateGenerator) Files() []util.File {
	return tg.files
}

func (tg *TemplateGenerator) SetFiles(files []util.File) {
	tg.files = files
}

func (tg *TemplateGenerator) Degenerate() error {
	outPath := tg.OutputDir
	if !filepath.IsAbs(outPath) {
		outPath, _ = filepath.Abs(outPath)
	}
	for path := range tg.Templates {
		abPath := filepath.Join(outPath, path)
		if err := os.RemoveAll(abPath); err != nil {
			return fmt.Errorf("remove file '%s' failed, err: %v", path, err.Error())
		}
	}
	for dir, exist := range tg.Dirs {
		if !exist {
			abDir := filepath.Join(outPath, dir)
			if err := os.RemoveAll(abDir); err != nil {
				return fmt.Errorf("remove directory '%s' failed, err: %v", dir, err.Error())
			}
		}
	}
	return nil
}
