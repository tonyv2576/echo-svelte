package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"
)

type FileManager struct {
	ProjectName string
	Port        string
	Typescript  bool
	Path        string
}

func NewFileManager(path string) (*FileManager, error) {
	fm := FileManager{
		Path: path,
	}

	return &fm, nil
}

func (fm *FileManager) Init(projectName string, typescript bool) error {
	fm.ProjectName = projectName
	fm.Typescript = typescript

	if err := fm.initModule(projectName); err != nil {
		return err
	}
	if err := fm.initVite(typescript); err != nil {
		return err
	}
	if err := fm.initPackages(); err != nil {
		return err
	}

	return nil
}

func (fm *FileManager) ApplyDefaults(port string) error {
	fm.Port = port
	if err := fm.applyMainDefault(); err != nil {
		return err
	}

	if err := fm.applyViteDefault(); err != nil {
		return err
	}

	if err := fm.goTidy(); err != nil {
		return err
	}

	if err := fm.viteBuild(); err != nil {
		return err
	}

	return nil
}

func (fm *FileManager) applyViteDefault() error {
	f, err := os.OpenFile(fm.configPath(), os.O_RDWR, os.ModeAppend)
	if err != nil {
		return doErr("failed to open '%v': %v", fm.configPath(), err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf(DEFAULT_VITE_CODE, fm.Port))
	if err != nil {
		return doErr("failed to update vite config: %v", err)
	}

	return fm.updatePackage()
}

func (fm *FileManager) applyMainDefault() error {
	f, err := os.Create(fm.mainPath())
	if err != nil {
		return doErr("failed to create '%v': %v", fm.mainPath(), err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf(DEFAULT_GO_CODE, fm.Port))
	if err != nil {
		return doErr("failed to update main.go: %v", err)
	}
	return nil
}

func (fm *FileManager) updatePackage() error {

	data, err := os.ReadFile(fm.packagePath())
	if err != nil {
		return doErr("failed to read package.json: %v", err)
	}

	var values map[string]any
	err = json.Unmarshal(data, &values)
	if err != nil {
		return doErr("failed to unmarshal package.json: %v", err)
	}

	// update values here
	if scr, ok := values["scripts"]; ok {
		if scripts, ok := scr.(map[string]any); ok {
			scripts["backend"] = "cd .. && air"
			scripts["dev:full"] = "concurrently \"npm run dev\" \"npm run backend\""
			result, err := json.Marshal(values)
			if err != nil {
				return doErr("failed to marshal package.json: %v", err)
			}

			f, err := os.OpenFile(fm.packagePath(), os.O_RDWR|os.O_TRUNC, os.ModeAppend)
			if err != nil {
				return doErr("failed to open '%v': %v", fm.configPath(), err)
			}
			defer f.Close()
			_, err = f.WriteString(string(result))
			if err != nil {
				return doErr("failed to update package.json: %v", err)
			}
			return nil
		}
	}

	return doErr("failed to update scripts parameter in package.json")
}

func (fm *FileManager) goTidy() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = fm.Path
	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("go mod tidy timed out")
		}
		return doErr("failed to run go mod tidy: %v", err.Error())
	}

	return nil
}

func (fm *FileManager) viteBuild() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npm", "run", "build", "--prefix", "frontend")
	cmd.Dir = fm.Path
	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("npm build timed out")
		}
		return doErr("failed to run npm build: %v", err)
	}

	return nil
}

func (fm *FileManager) initPackages() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "get", "github.com/labstack/echo/v4@latest")
	cmd.Dir = fm.Path
	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("echo installation timed out")
		}
		return doErr("failed to get echo package: %v", err.Error())
	}

	cmd = exec.CommandContext(ctx, "go", "get", "github.com/labstack/echo/v4/middleware@latest")
	cmd.Dir = fm.Path
	if err := cmd.Run(); err != nil {

		if err == context.DeadlineExceeded {
			return doErr("echo middleware installation timed out")
		}
		return doErr("failed to get echo middleware package: %v", err.Error())
	}

	cmd = exec.CommandContext(ctx, "go", "install", "github.com/air-verse/air@latest")
	cmd.Dir = fm.Path
	if err := cmd.Run(); err != nil {

		if err == context.DeadlineExceeded {
			return doErr("air installation timed out")
		}
		return doErr("failed to install air package: %v", err.Error())
	}

	return nil
}

func (fm *FileManager) initModule(projectName string) error {
	if len(projectName) <= 0 {
		return doErr("go module name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "mod", "init", projectName)
	cmd.Dir = fm.Path

	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("go module setup timed out")
		}
		return doErr("failed to initialize go module: %v", err.Error())
	}

	return nil
}

func (fm *FileManager) initVite(typescript bool) error {
	tmplName := "svelte"
	if typescript {
		tmplName += "-ts"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npm", "create", "vite@latest", "frontend", "--", "--template", tmplName)
	cmd.Dir = fm.Path

	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("vite setup timed out")
		}
		return doErr("failed to setup vite frontend: %v", err.Error())
	}

	cmd = exec.CommandContext(ctx, "npm", "install", "-D", "concurrently")
	cmd.Dir = fm.frontendPath()

	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("concurrently installation timed out")
		}
		return doErr("failed to install concurrently package: %v", err.Error())
	}

	cmd = exec.CommandContext(ctx, "npm", "install")
	cmd.Dir = fm.frontendPath()

	if err := cmd.Run(); err != nil {
		if err == context.DeadlineExceeded {
			return doErr("frontend installation timed out")
		}
		return doErr("failed to install vite frontend: %v", err.Error())
	}

	return nil
}

func (fm *FileManager) frontendPath() string {
	return path.Join(fm.Path, "frontend")
}

func (fm *FileManager) configPath() string {
	return path.Join(fm.frontendPath(), "vite.config.js")
}
func (fm *FileManager) packagePath() string {
	return path.Join(fm.frontendPath(), "package.json")
}

func (fm *FileManager) mainPath() string {
	return path.Join(fm.Path, "main.go")
}
