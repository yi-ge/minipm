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

func runCmd(c *Command) {
	log.Printf("Starting process: %s\n", c.Name)

	for {
		logFilePath := filepath.Join(os.Getenv("HOME"), ".minipm", "logs", fmt.Sprintf("%s-%s.log", c.Name, time.Now().Format("2006-01-02_15-04-05")))
		newCmd := fmt.Sprintf("nohup %s >> %s 2>&1 &", c.Cmd, logFilePath)
		fmt.Printf("Starting process '%s'...\n", c.Name)
		err := exec.Command("sh", "-c", newCmd).Start()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		time.Sleep(time.Second * 2)

		if !c.IsRunning() {
			log.Printf("Process %s exited immediately after start\n", c.Name)
			continue
		}

		log.Printf("Process %s started successfully\n", c.Name)
		break
	}
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

			go runCmd(cmd)
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

func latestLogFilePath(name string) string {
	logDir := filepath.Join(os.Getenv("HOME"), ".minipm", "logs")
	logFiles, err := filepath.Glob(filepath.Join(logDir, fmt.Sprintf("%s-*.log", name)))
	if err != nil || len(logFiles) == 0 {
		return ""
	}
	latestLogFile := logFiles[0]
	latestLogTime := time.Time{}
	for _, logFile := range logFiles {
		fi, err := os.Stat(logFile)
		if err != nil {
			continue
		}
		if fi.ModTime().After(latestLogTime) {
			latestLogTime = fi.ModTime()
			latestLogFile = logFile
		}
	}
	return latestLogFile
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
			Cmd:  strings.TrimSpace(line),
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

		logsDir := filepath.Join(os.Getenv("HOME"), ".minipm", "logs")
		err = os.Mkdir(logsDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create .minipm/logs directory: %s", err)
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
			fmt.Println("Usage: minipm [command] [options]")
			fmt.Println()
			fmt.Println("Commands:")
			fmt.Println("  run <command>      Start a new process and add it to the process list")
			fmt.Println("  list               List all managed processes")
			fmt.Println("  --enable           Register the program as a service")
			fmt.Println("  --start            Start the program as a service")
			fmt.Println("  --stop             Stop the program service")
			fmt.Println("  --disable          Unregister the program as a service")
			fmt.Println("  -v, --version      Display the version number")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  -h, --help         Display this help message")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  minipm run \"python3 myscript.py\"")
			fmt.Println("  minipm list")
			fmt.Println("  minipm --enable")
			fmt.Println("  minipm --start")
			fmt.Println("  minipm --stop")
			fmt.Println("  minipm --disable")
			os.Exit(0)
		case "run":
			if len(os.Args) < 3 {
				fmt.Println("Usage: minipm run <command>")
				os.Exit(1)
			}

			pmListPath := filepath.Join(os.Getenv("HOME"), ".minipm", "pm-list.txt")
			f, err := os.OpenFile(pmListPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalf("Failed to open pm list file %s: %s", pmListPath, err)
			}
			defer f.Close()

			cmd := &Command{
				Name: filepath.Base(os.Args[2]),
				Cmd:  strings.Join(os.Args[2:], " "),
			}

			fmt.Fprintln(f, cmd.Cmd)

			runCmd(cmd)

			os.Exit(0)
		case "list":
			pmListPath := filepath.Join(os.Getenv("HOME"), ".minipm", "pm-list.txt")
			pmList, err := loadPmList(pmListPath)
			if err != nil {
				log.Fatalf("Failed to load pm list: %s", err)
			}

			for _, cmd := range pmList {
				logPath := latestLogFilePath(cmd.Name)
				fmt.Printf("%s: %s\n", cmd.Cmd, logPath)
			}

			os.Exit(0)
		case "-v", "--version":
			fmt.Println("minipm version 0.1.0")
			os.Exit(0)
		case "--enable":
			err := s.Install()
			if err != nil {
				fmt.Printf("Failed to register service: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("Service registered successfully")
			os.Exit(0)
		case "--start":
			err := s.Start()
			if err != nil {
				fmt.Printf("Failed to start service: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("Service started successfully")
			os.Exit(0)
		case "--stop":
			err := s.Stop()
			if err != nil {
				fmt.Printf("Failed to stop service: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("Service stopped successfully")
			os.Exit(0)
		case "--disable":
			err = s.Uninstall()
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			fmt.Println("Service uninstalled")
		default:
			fmt.Printf("Error: Invalid command '%s'\n", os.Args[1])
			os.Exit(1)
		}
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
