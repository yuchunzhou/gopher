package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	Path "path"
	"path/filepath"
	"sync"

	"github.com/urfave/cli/v2"
)

func createProject(path string) {
	// 判断目录是否存在
	_, err := os.Stat(path)
	if err == nil {
		fmt.Printf("Destination `%s` already exists\n", path)
		return
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Println("Create project failed:", err)
		return
	}

	// 初始化项目
	main := `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`
	mainPath := filepath.Join(path, "main.go")
	mainFile, err := os.Create(mainPath)
	if err != nil {
		fmt.Println("Create project failed:", err)
		return
	}

	_, err = mainFile.WriteString(main)
	if err != nil {
		fmt.Println("Create project failed:", err)
		return
	}

	base := Path.Base(path)
	path, err = filepath.Abs(path)
	if err != nil {
		fmt.Println("Create project failed:", err)
		return
	}

	cmd := exec.Command("go", "mod", "init", base)
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		fmt.Println("Create project failed:", err)
		return
	}

	cmd = exec.Command("git", "init")
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		fmt.Println("Create project failed:", err)
		return
	}

	fmt.Println("Created Golang project", base)
}

func updateProject(path string) {
	cmd := exec.Command("git", "config", "pull.rebase", "false")
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		fmt.Println("Update project failed:", err)
		return
	}

	base := Path.Base(path)
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd = exec.Command("git", "pull", "--all")
	cmd.Dir = path
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(string(stdout.Bytes()))
		fmt.Println(string(stderr.Bytes()))
		fmt.Printf("Update project %s failed\n", base)
		return
	}

	fmt.Println("Updated project", base)
}

func main() {
	app := cli.NewApp()
	app.Usage = "A gopher cli tool"
	app.Commands = []*cli.Command{
		{
			Name:  "new",
			Usage: "Create new Golang projects",
			Action: func(c *cli.Context) error {
				var wg sync.WaitGroup
				for _, path := range c.Args().Slice() {
					wg.Add(1)
					go func(path string) {
						defer wg.Done()
						createProject(path)
					}(path)
				}

				wg.Wait()
				return nil
			},
		},
		{
			Name:  "update",
			Usage: "Update git projects",
			Action: func(c *cli.Context) error {
				var wg sync.WaitGroup
				if c.Args().Len() == 0 {
					files, err := os.ReadDir(".")
					if err != nil {
						fmt.Println("Update projects failed:", err)
						return nil
					}

					paths := make([]string, 0)
					for _, file := range files {
						if file.IsDir() {
							path, _ := filepath.Abs(file.Name())
							git := filepath.Join(path, ".git")
							_, err = os.Stat(git)
							if err == nil {
								paths = append(paths, path)
							}
						}
					}

					fmt.Println(paths)
					for _, path := range paths {
						wg.Add(1)
						go func(path string) {
							defer wg.Done()
							updateProject(path)
						}(path)
					}
				} else {
					for _, path := range c.Args().Slice() {
						wg.Add(1)
						go func(path string) {
							defer wg.Done()
							updateProject(path)
						}(path)
					}
				}

				wg.Wait()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
