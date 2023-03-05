package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/service"
)

type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})

	go p.run()

	return nil
}

func (p *program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}

func (p *program) run() {
	for {
		pmListPath := filepath.Join(os.Getenv("HOME"), ".minipm", "pm-list.txt")
		pmList, err := loadPmList(pmListPath)
		if err != nil {
			log.Printf("Failed to load pm list: %s\n", err)
		}

		for _, cmd := range pmList {
			if cmd.IsRunning() {
				continue
			}

			go func(c *Command) {
				log.Printf("Starting process: %s\n", c.Name)

				for {
					err := c.Start()
					if err != nil {
						log.Printf("Failed to start process %s: %s\n", c.Name, err)
						time.Sleep(time.Second * 5)
						continue
					}

					time.Sleep(time.Second * 2)

					if !c.IsRunning() {
						log.Printf("Process %s exited immediately after start\n", c.Name)
						continue
					}

					log.Printf("Process %s started successfully\n", c.Name)
					break
				}
			}(cmd)
		}

		select {
		case <-p.exit:
			return
		case <-time.After(time.Second * 30):
		}
	}
}

type Command struct {
	Name string
	Cmd  string
}

func (c *Command) IsRunning() bool {
	out, _ := exec.Command("pgrep", "-f", c.Cmd).Output()
	return len(out) > 0
}

func (c *Command) Start() error {
	logPath := filepath.Join(os.Getenv("HOME"), ".minipm", "logs", fmt.Sprintf("%s-%s.log", c.Name, time.Now().Format("2006-01-02_15-04-05")))
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %s", logPath, err)
	}
	defer f.Close()

	cmd := exec.Command("sh", "-c", c.Cmd)
	cmd.Stdout = f
	cmd.Stderr = f

	return cmd.Start()
}

func loadPmList(path string) ([]*Command, error) {
	var pmList []*Command

	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，不输出任何内容
			return pmList, nil
		}
		return pmList, fmt.Errorf("failed to open pm list file %s: %s", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		name := strings.Split(line, " ")[0]
		cmd := &Command{
			Name: name,
			Cmd:  strings.TrimSpace(strings.Replace(line, name, "", 1)),
		}

		pmList = append(pmList, cmd)
	}

	return pmList, scanner.Err()
}

func main() {
	// 检查并创建 .minipm 文件夹
	pmDir := filepath.Join(os.Getenv("HOME"), ".minipm")
	if _, err := os.Stat(pmDir); os.IsNotExist(err) {
		err = os.Mkdir(pmDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create .minipm directory: %s", err)
		}
	}

	svcConfig := &service.Config{
		Name:        "minipm",
		DisplayName: "Mini Process Manager",
		Description: "A simple process manager for Linux",
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Println("Usage: minipm [command]")
			fmt.Println()
			fmt.Println("Commands:")
			fmt.Println("  start <command>  Start a new process and add it to the process list")
			fmt.Println("  list             List all managed processes")
			fmt.Println("  -v, --version    Display the version number")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  -h, --help       Display this help message")
			os.Exit(0)
		case "start":
			if len(os.Args) < 3 {
				fmt.Println("Usage: minipm start <command>")
				os.Exit(1)
			}

			pmListPath := filepath.Join(os.Getenv("HOME"), ".minipm", "pm-list.txt")
			f, err := os.OpenFile(pmListPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalf("Failed to open pm list file %s: %s", pmListPath, err)
			}
			defer f.Close()

			cmd := strings.Join(os.Args[2:], " ")
			fmt.Fprintf(f, "%s %s\n", filepath.Base(os.Args[2]), cmd)

			fmt.Printf("Started process: %s\n", cmd)
			os.Exit(0)
		case "list":
			pmListPath := filepath.Join(os.Getenv("HOME"), ".minipm", "pm-list.txt")
			pmList, err := loadPmList(pmListPath)
			if err != nil {
				log.Fatalf("Failed to load pm list: %s", err)
			}

			for _, cmd := range pmList {
				logPath := filepath.Join(os.Getenv("HOME"), ".minipm", "logs", fmt.Sprintf("%s-%s.log", cmd.Name, time.Now().Format("2006-01-02")))
				fmt.Printf("%s: %s\n", cmd.Cmd, logPath)
			}

			os.Exit(0)
		case "-v", "--version":
			fmt.Println("minipm version 0.1.0")
			os.Exit(0)
		}
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
