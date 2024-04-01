# minipm

`minipm` is a simple process manager written in Go, allowing you to start, stop, and daemonize processes on Linux and automatically restart them to maintain system stability. minipm also provides additional features such as registering services and outputting process lists.

## Installation

First, make sure you have installed Go and Git. Then, clone the minipm repository and install the program with the following commands:

```bash

```

## Usage

To start minipm, run the following command in your terminal:

```bash
minipm
```

This will start the minipm service and wait for input. To add a new process, use the following command:

```bash
minipm run <command>
```

This will add a new process and set its command-line arguments to <command>. minipm will automatically start the process and daemonize it. If the process terminates unexpectedly, minipm will automatically restart it.

To view all currently managed processes, use the following command:

```bash
minipm list
```

This will output a process list containing all currently managed processes and their statuses.

To register the minipm service, use the following command:

```bash
minipm --enable
```

This will register the minipm service to automatically start on system boot. To start the service, use the following command:

```bash
minipm --start
```

This will start the minipm service. To stop the service, use the following command:

```bash
minipm --stop
```

This will stop the minipm service.

## Version

To check the version of minipm, use the following command:

```bash
minipm --version
```

To get help with minipm, use the following command:

```bash
minipm --help
```

## Author

minipm was developed by ChatGPT and Yige. If you have any suggestions or questions, please feel free to contact us.

## License

minipm is licensed under the MIT License. See the LICENSE file for details.
