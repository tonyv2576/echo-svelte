package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type ProjectConfig struct {
	Name       string
	Directory  string
	Typescript bool

	Port    int
	DevMode bool

	Files *FileManager
}

func (p *ProjectConfig) checkForRights() error {
	if err := exec.Command("cd").Run(); err != nil {
		return doErr("application requires elevated rights")
	}

	return nil
}

func (p *ProjectConfig) RunDevMode() error {
	fmt.Println("Running in dev mode...")
	loc, err := os.Getwd()
	if err != nil {
		return doErr("unable to get current path")
	}
	p.Directory = loc
	p.Name = path.Base(p.Directory)

	cmd := exec.Command("npm", "run", "dev:full")
	cmd.Dir = path.Join(p.Directory, "frontend")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return doErr("failed to start dev command: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		return doErr("error while running dev command: %v", err)
	}

	return nil
}

func (p *ProjectConfig) CreateDirectory() error {
	fmt.Println("Creating directory...")

	err := os.MkdirAll(p.Directory, os.ModePerm)
	if err != nil {
		if err == os.ErrExist {
			fmt.Println("[WARNING] Directory already exists")
			return p.ValidatePath()
		}
	}

	fileMan, err := NewFileManager(p.Directory)
	if err != nil {
		return err
	}
	p.Files = fileMan

	return nil
}

func (p *ProjectConfig) InstallModule() error {
	fmt.Println("Installing frontend...")
	if p.Files == nil {
		return doErr("file manager is uninitialized")
	}
	return p.Files.Init(p.Name, p.Typescript)
}

func (p *ProjectConfig) ApplyDefaults() error {
	fmt.Println("Applying defaults...")
	if p.Files == nil {
		return doErr("file manager is uninitialized")
	}
	return p.Files.ApplyDefaults(fmt.Sprintf(":%4.v", p.Port))
}

func (p *ProjectConfig) Validate() error {
	if len(p.Name) <= 0 {
		return doErr("project name is required")
	}

	fmt.Println("Validating project path...")
	if len(p.Directory) <= 0 || strings.HasPrefix(p.Directory, ".") {
		loc, err := os.Getwd()
		if err != nil {
			return doErr("unable to get current path")
		}
		for strings.HasPrefix(p.Directory, "../") {
			p.Directory = strings.TrimPrefix(p.Directory, "../")
			loc = path.Dir(loc)
		}
		if len(p.Directory) <= 0 {
			loc = path.Join(loc, p.Name)
		} else if p.Directory != "./" {
			loc = path.Join(loc, p.Directory)
		}
		p.Directory = loc
		return p.CreateDirectory()
	}

	return p.ValidatePath()
}
func (p *ProjectConfig) ValidatePath() error {
	// validate path
	dir, err := os.ReadDir(p.Directory)
	if err != nil {
		return doErr("the path %v is invalid", p.Directory)
	}

	// make sure path is empty
	if len(dir) > 0 {
		return doErr("the project directory must be empty")
	}
	return nil
}

func doErr(val string, args ...any) error {
	return fmt.Errorf(val, args...)
}
