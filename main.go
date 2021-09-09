package main

import (
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

func main() {
	app := cli.NewApp()
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
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
