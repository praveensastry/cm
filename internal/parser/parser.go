package parser

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gotree "github.com/DiSiqueira/GoTree"
	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/olekukonko/tablewriter"
	"github.com/praveensastry/cm/terminal"

	"gopkg.in/ini.v1"
)

type SpecList struct {
	Specs map[string]*Spec
}

type Spec struct {
	Version  string   `ini:"VERSION"`
	Requires []string `ini:"REQUIRES,omitempty"`
	Packages Packages `ini:"PACKAGES"`
	Configs  Configs  `ini:"CONFIGS"`
	Content  Content  `ini:"CONTENT"`
	Commands Commands `ini:"COMMANDS"`
	SpecFile string   `ini:"-"`
	SpecRoot string   `ini:"-"`
}

type Packages struct {
	AptGet       []string `ini:"apt_get"`
	SkipPackages bool     `ini:"skip_packages"`
}

type Configs struct {
	DebianRoot      string `ini:"debian_root"`
	SkipInterpolate bool   `ini:"skip_interpolate"`
}

type Content struct {
	Source     string `ini:"source"`
	DebianRoot string `ini:"debian_root"`
}

type Commands struct {
	Pre      []string `ini:"pre,omitempty"`
	Post     []string `ini:"post,omitempty"`
	SkipPre  bool     `ini:"skip_pre"`
	SkipPost bool     `ini:"skip_post"`
	TailPre  bool     `ini:"tail_pre"`
	TailPost bool     `ini:"tail_post"`
}

type SpecSummary struct {
	Name      string
	Requires  []string
	PreCmds   []string
	AptCmds   []string
	Transfers *FileTransfers
	PostCmds  []string
}

type FileTransfer struct {
	Source      string
	Destination string
	Folder      string
	Chown       string
	Chmod       string
	Interpolate bool
}

// Jobs that run locally
type LocalJob struct {
	Class       string
	Sequence    string
	Locale      string
	Deltas      chan string
	Notices     chan string
	Responses   chan string
	Information chan string
	Errors      chan error
	SpecName    string
	SpecList    *SpecList
	WaitGroup   *sync.WaitGroup
}

type FileTransfers []FileTransfer

// Reads in all the specs and builds a SpecList
func GetSpecs() (*SpecList, error) {

	var err error
	specList := new(SpecList)
	specList.Specs = make(map[string]*Spec)

	currentUser, _ := user.Current()
	// To-Do add support for remote specs
	candidates := []string{
		currentUser.HomeDir + "/.cmspecs/",
		"./specs/",
	}

	walkFn := func(path string, fileInfo os.FileInfo, inErr error) (err error) {
		if inErr == nil && !fileInfo.IsDir() && strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".spec") {
			err = specList.scanFile(path)
		}
		return
	}

	// Walk each of the candidate folders
	for _, folder := range candidates {
		err = filepath.Walk(folder, walkFn)
	}

	return specList, err
}

// Scans a given file and if it is a spec, adds it to the spec list
func (s *SpecList) scanFile(file string) error {
	// This is most likely a spec file, so lets try to pull a struct from it

	cfg, err := ini.Load(file)
	if err != nil {
		return err
	}

	specName := cfg.Section("").Key("NAME").String()

	spec := new(Spec)

	if len(specName) > 0 {
		err := cfg.MapTo(spec)
		if err != nil {
			return err
		}
		spec.SpecFile = file
		spec.SpecRoot = path.Dir(file)
		s.Specs[specName] = spec
	}

	return nil
}

// Checks if a given spec exists
func (s *SpecList) SpecExists(spec string) bool {
	if _, ok := s.Specs[spec]; ok {
		return true
	}
	return false
}

// Returns the apt-get commands for a given spec
func (s *SpecList) AptGetCmds(specName string) (cmds []string) {
	packages := s.getAptPackages(specName)
	if len(packages) > 0 {
		cmds = []string{"sudo apt-get update -o Dpkg::Options::=\"--force-confdef\" -o Dpkg::Options::=\"--force-confold\"", "sudo apt-get install -y -f --assume-yes --allow-unauthenticated " + strings.Join(packages, " ")}
	}

	return cmds
}

// Returns the pre-configure commands
func (s *SpecList) PreCmds(specName string) []string {
	return s.getPreCommands(specName)
}

// Returns the requires
func (s *SpecList) Requires(specName string) []string {
	requires := s.getRequires(specName)
	if requires != nil {
		return strings.Split(requires.Print(), "\n")
	}

	return nil
}

// Returns the post-configure commands
func (s *SpecList) PostCmds(specName string) []string {
	return s.getPostCommands(specName)
}

func (s *SpecList) DebianFileTransferList(specName string) *FileTransfers {

	fileTransfers := new(FileTransfers)

	fileTransfers = s.getDebianFileTransfers(specName)
	return fileTransfers

}

// Recursive unexported func for FileTransferList
func (s *SpecList) getDebianFileTransfers(specName string) *FileTransfers {

	// The requested spec
	spec := s.Specs[specName]
	files := new(FileTransfers)

	if spec == nil {
		return files
	}

	srcConfFolder := spec.SpecRoot + "/configs/"
	destConfFolder := spec.Configs.DebianRoot
	interpolate := true
	if spec.Configs.SkipInterpolate == true {
		interpolate = false
	}

	if spec.Configs.DebianRoot != "" {
		// Walk the Configs folder and append each file
		walkFn := func(path string, fileInfo os.FileInfo, inErr error) (err error) {
			if inErr == nil && !fileInfo.IsDir() {
				destination := destConfFolder + strings.TrimPrefix(path, srcConfFolder)
				files.add(FileTransfer{
					Source:      path,
					Destination: destination,
					Folder:      filepath.Dir(destination),
					Interpolate: interpolate,
				})
			}
			return
		}
		filepath.Walk(srcConfFolder, walkFn)
	}

	srcContentFolder := spec.SpecRoot + "/content/"
	destContentFolder := spec.Content.DebianRoot

	if spec.Content.DebianRoot != "" && spec.Content.Source == "spec" {
		// Walk the Configs folder and append each file
		walkFn := func(path string, fileInfo os.FileInfo, inErr error) (err error) {
			if inErr == nil && !fileInfo.IsDir() {
				destination := destContentFolder + strings.TrimPrefix(path, srcContentFolder)
				files.add(FileTransfer{
					Source:      path,
					Destination: destination,
					Folder:      filepath.Dir(destination),
				})
			}
			return
		}
		filepath.Walk(srcContentFolder, walkFn)
	}

	for _, reqSpec := range spec.Requires {
		reqFiles := s.getDebianFileTransfers(reqSpec)
		*files = append(*files, *reqFiles...)
	}

	return files
}

func (f *FileTransfers) add(file FileTransfer) {
	*f = append(*f, file)
}

func (s *SpecList) ShowSpecBuild(specName string) {

	terminal.PrintAnsi(SpecBuildTemplate, SpecSummary{
		Name:      specName,
		Requires:  s.Requires(specName),
		PreCmds:   s.PreCmds(specName),
		AptCmds:   s.AptGetCmds(specName),
		Transfers: s.DebianFileTransferList(specName),
		PostCmds:  s.PostCmds(specName),
	})
}

var SpecBuildTemplate = `
{{ansi ""}}{{ ansi "underscore"}}{{ ansi "bright" }}{{ ansi "fgwhite"}}[{{ .Name }}]{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}                Requires: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range .Requires }}{{ . }}
				  {{ end }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}  pre-configure Commands: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range .PreCmds }}{{ . }}
				  {{ end }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}        apt-get Commands: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range .AptCmds }}{{ . }}
				  {{ end }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}          File Transfers: {{ ansi ""}}{{ ansi "fgcyan"}}{{range .Transfers}}
				      Source: {{ .Source }}
				 Destination: {{ .Destination }}
				      Folder: {{ .Folder }}
				 {{ end }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}} post-configure Commands: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range .PostCmds }}{{ . }}
				  {{ end }}{{ ansi ""}}
`

// Run Local configuration on this machine
func (s *SpecList) LocalConfigure(specName, class, sequence, locale string) {

	var wg sync.WaitGroup
	wg.Add(1)

	deltas := make(chan string, 10)
	notices := make(chan string, 10)
	responses := make(chan string, 10)
	information := make(chan string, 10)
	errors := make(chan error, 10)

	job := LocalJob{
		Deltas:      deltas,
		Notices:     notices,
		Responses:   responses,
		Information: information,
		Errors:      errors,
		SpecName:    specName,
		SpecList:    s,
		WaitGroup:   &wg,
		Class:       class,
		Sequence:    sequence,
		Locale:      locale,
	}

	// Launch it!
	go job.Run()

	// Display Output of Job
	go func() {
		for {
			select {
			case delta := <-deltas:
				terminal.Delta(delta)
			case notice := <-notices:
				terminal.Notice(notice)
			case resp := <-responses:
				terminal.Response(resp)
			case info := <-information:
				terminal.Information(info)
			case err := <-errors:
				terminal.ErrorLine(err.Error())
			}
		}
	}()

	wg.Wait()

	time.Sleep(time.Second)
}

func (job *LocalJob) Run() {
	defer job.WaitGroup.Done()

	// Run pre configure commands
	preCmds := job.SpecList.PreCmds(job.SpecName)
	for _, preCmd := range preCmds {
		job.Deltas <- "Running pre-configuration command: [" + preCmd + "]"
		_, err := job.runCommand(preCmd, "pre-configuration")
		if err != nil {
			job.Errors <- err
			job.Errors <- errors.New("pre-configuration command: [" + preCmd + "] Failed! Aborting futher tasks for this server..")
			return
		}
		job.Information <- "pre-configuration command: [" + preCmd + "] Succeeded!"
	}

	// Run Apt-Get Commands
	aptCmds := job.SpecList.AptGetCmds(job.SpecName)
	for _, aptCmd := range aptCmds {
		job.Deltas <- "Running apt-get command: [" + aptCmd + "]"
		_, err := job.runCommand(aptCmd, "apt-get")
		if err != nil {
			job.Errors <- err
			job.Errors <- errors.New("apt-get command: [" + aptCmd + "] Failed! Aborting futher tasks for this server..")
			return
		}
		job.Information <- "apt-get command: [" + aptCmd + "] Succeeded!"
	}

	// Transfer any files we need to transfer
	fileList := job.SpecList.DebianFileTransferList(job.SpecName)
	if len(*fileList) > 0 {
		job.Deltas <- "Starting file copy..."
		err := job.transferFiles(fileList, "Configuration and Content Files")
		if err != nil {
			job.Errors <- errors.New("File Copy Failed! Aborting futher tasks for this server..")
			return
		}
		job.Information <- "File Copy Succeeded!"
	}

	// Run post configure commands
	postCmds := job.SpecList.PostCmds(job.SpecName)
	for _, postCmd := range postCmds {
		job.Deltas <- "Running post-configuration command: [" + postCmd + "]"
		_, err := job.runCommand(postCmd, "post-configuration")
		if err != nil {
			job.Errors <- err
			job.Errors <- errors.New("post-configuration command: [" + postCmd + "] Failed!")
		} else {
			job.Information <- "post-configuration command: [" + postCmd + "] Succeeded!"
		}
	}
}

func (j *LocalJob) runCommand(command string, name string) (string, error) {

	if len(command) > 0 {
		parts := strings.Fields(command)
		cmd := exec.Command(parts[0], parts[1:]...)

		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		err := cmd.Run()

		// TODO handle more verbose output, maybe from a verbose cli flag
		if err != nil {
			j.Responses <- stdoutBuf.String()
			j.Responses <- stderrBuf.String()
		}

		return stdoutBuf.String(), err
	}

	return "empty command!", nil

}

func (j *LocalJob) transferFiles(fileList *FileTransfers, name string) error {

	// Defer cleanup
	defer j.runCommand("sudo rm -rf /tmp/cm/*", "")

	for _, file := range *fileList {

		// Make our temp folder
		j.runCommand("mkdir -p /tmp/cm/"+file.Folder, "")
		_, err := j.runCommand("sudo mkdir -p "+file.Folder, "") // should prob add chown and chmod to the config structs to set it afterwards
		if err != nil {
			j.Errors <- errors.New("Unable to make directory: " + file.Folder)
			return err
		}

		j.Responses <- "Copying file: " + file.Destination

		// Read the file
		////////////////..........
		rf, err := os.Open(file.Source)
		defer rf.Close()
		if err != nil {
			j.Errors <- errors.New("Unable to open file: " + file.Source)
			return err
		}

		rfi, err := rf.Stat()
		if err != nil {
			j.Errors <- errors.New("Unable to inspect file: " + file.Source)
			return err
		}

		fileSize := rfi.Size()
		fileBytes := make([]byte, fileSize)

		_, err = rf.Read(fileBytes)
		if err != nil {
			j.Errors <- errors.New("Unable to read file: " + file.Source)
			return err
		}

		var outputFile []byte

		if file.Interpolate {

			j.Deltas <- "Interpolating on file: " + file.Destination

			// Interpolate
			////////////////..........
			tree, err := hil.Parse(string(fileBytes))
			if err != nil {
				j.Errors <- errors.New("Unable to parse file: " + file.Source)
				return err
			}

			config := &hil.EvalConfig{
				GlobalScope: &ast.BasicScope{
					VarMap: map[string]ast.Variable{
						"var.class": ast.Variable{
							Type:  ast.TypeString,
							Value: j.Class,
						},
						"var.sequence": ast.Variable{
							Type:  ast.TypeString,
							Value: j.Sequence,
						},
						"var.locale": ast.Variable{
							Type:  ast.TypeString,
							Value: j.Locale,
						},
						"var.specname": ast.Variable{
							Type:  ast.TypeString,
							Value: j.SpecName,
						},
					},
				},
			}

			result, err := hil.Eval(tree, config)
			if err != nil {
				j.Errors <- errors.New("Unable to evaluate file: " + file.Source)
				return err
			}

			outputFile = []byte(result.Value.(string))
		} else {

			j.Notices <- "Skipping Interpolation on file: " + file.Destination

			outputFile = fileBytes
		}

		// Write the file
		////////////////..........
		rf, err = os.Create("/tmp/cm" + file.Destination)
		defer rf.Close()

		if err != nil {
			j.Errors <- errors.New("Unable to create file: " + file.Destination)
			return err
		}
		if _, err := rf.Write(outputFile); err != nil {
			j.Errors <- errors.New("Unable to write file: " + file.Destination)
			return err
		}

		// mv
		j.runCommand("sudo mv /tmp/cm"+file.Destination+" "+file.Destination, "")

		j.Information <- "Completed copy of file: " + file.Destination
	}

	return nil

}

// Prints table of all available specs in a table
func (s *SpecList) PrintSpecInformation() {
	terminal.PrintAnsi(SpecTemplate, s)
}

var SpecTemplate = `{{range $name, $spec := .Specs}}
{{ansi ""}}{{ ansi "underscore"}}{{ ansi "bright" }}{{ ansi "fgwhite"}}[{{ $name }}]{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}                Version: {{ ansi ""}}{{ ansi "fgcyan"}}{{ $spec.Version }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}                   Root: {{ ansi ""}}{{ ansi "fgcyan"}}{{ $spec.SpecRoot }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}                   File: {{ ansi ""}}{{ ansi "fgcyan"}}{{ $spec.SpecFile }}{{ ansi ""}}

	{{ ansi "bright"}}{{ ansi "fgwhite"}}               Requires: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range $spec.Requires }}{{ . }} {{ end }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}           Apt Packages: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range $spec.Packages.AptGet }}{{ . }} {{ end }}{{ ansi ""}}

	{{ ansi "bright"}}{{ ansi "fgwhite"}}    Debian Configs Root: {{ ansi ""}}{{ ansi "fgcyan"}}{{ $spec.Configs.DebianRoot }}{{ ansi ""}}

	{{ ansi "bright"}}{{ ansi "fgwhite"}}         Content Source: {{ ansi ""}}{{ ansi "fgcyan"}}{{ $spec.Content.Source }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}}    Debian Content Root: {{ ansi ""}}{{ ansi "fgcyan"}}{{ $spec.Content.DebianRoot }}{{ ansi ""}}

	{{ ansi "bright"}}{{ ansi "fgwhite"}}  Pre-configure command: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range $spec.Commands.Pre }}{{ printf "%s" . }}
				 {{ end }}{{ ansi ""}}
	{{ ansi "bright"}}{{ ansi "fgwhite"}} Post-configure command: {{ ansi ""}}{{ ansi "fgcyan"}}{{ range $spec.Commands.Post }}{{ printf "%s" . }}
				 {{ end }}{{ ansi ""}}


{{ ansi "fgwhite"}}------------------------------------------------------------------------------------------------
{{ ansi ""}}
{{ end }}
`

// Recursive unexported func for PreCmds
func (s *SpecList) getPreCommands(specName string) []string {
	// The requested spec
	spec := s.Specs[specName]
	var commands []string
	if spec == nil || spec.Commands.SkipPre {
		return nil
	}

	// gather all required pre configure commands for this spec
	for _, pre := range spec.Commands.Pre {
		if pre != "" {
			commands = append(commands, pre)
		}
	}

	// Loop through this specs requirements to all other pre configure commands we need
	for _, reqSpec := range spec.Requires {
		if reqSpec != "" {
			commands = append(s.getPreCommands(reqSpec), commands...) // prepend
		}
	}

	// Dedupe, remove later ones
	for index := 0; index < len(commands); index++ {
		for compare := index + 1; compare < len(commands); compare++ {
			if commands[index] == commands[compare] {
				commands = append(commands[:compare], commands[compare+1:]...)
				compare--
			}
		}
	}

	return commands
}

func (s *SpecList) getRequires(specName string) gotree.Tree {
	// The requested spec
	spec := s.Specs[specName]
	requires := gotree.New(specName)

	if spec == nil {
		return requires
	}

	// gather all requires for this spec
	for _, req := range spec.Requires {
		if req != "" && req != "\"\"" {
			requires.Add(req)

			subreqs := s.getRequires(req)
			if len(subreqs.Items()) > 0 {
				requires.AddTree(subreqs)
			}
		}
	}

	return requires
}

// Recursive unexported func for PostCmds
func (s *SpecList) getPostCommands(specName string) []string {
	// The requested spec
	spec := s.Specs[specName]
	var commands []string

	if spec == nil || spec.Commands.SkipPost {
		return nil
	}

	// gather all required post configure commands for this spec
	if len(spec.Commands.Post) > 0 {
		for _, post := range spec.Commands.Post {
			commands = append(commands, post)
		}
	}

	// Loop through this specs requirements to all other post configure commands we need
	for _, reqSpec := range spec.Requires {
		commands = append(s.getPostCommands(reqSpec), commands...) // prepend
		//commands = append(commands, s.getPostCommands(reqSpec)...) // append
	}

	// Dedupe, remove later ones
	for index := 0; index < len(commands); index++ {
		for compare := index + 1; compare < len(commands); compare++ {
			if commands[index] == commands[compare] {
				commands = append(commands[:compare], commands[compare+1:]...)
				compare--
			}
		}
	}

	return commands
}

// Recursive unexported func for AptGetCmds
func (s *SpecList) getAptPackages(specName string) []string {
	// The requested spec
	spec := s.Specs[specName]
	var packages []string
	if spec == nil || spec.Packages.SkipPackages {
		return nil
	}

	// Gather all required apt-get packages for this spec
	packages = append(packages, spec.Packages.AptGet...)

	// Loop through this specs requirements gather to all other apt-get packages we need
	for _, reqSpec := range spec.Requires {
		packages = append(packages, s.getAptPackages(reqSpec)...)
	}

	// Dedupe
	for index := 0; index < len(packages); index++ {
		for compare := index + 1; compare < len(packages); compare++ {
			if packages[index] == packages[compare] {
				packages = append(packages[:compare], packages[compare+1:]...)
				compare--
			}
		}
	}

	return packages
}

func printTable(header []string, rows [][]string) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

func addSpaces(s string, w int) string {
	if len(s) < w {
		s += strings.Repeat(" ", w-len(s))
	}
	return s
}
