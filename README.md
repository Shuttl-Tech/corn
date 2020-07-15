# Corn

Corn is a simple command line application to execute periodic tasks.

## Installing

Corn binaries are available for Linux and OSX platform on github release page. Or, if you prefer, it can be built from source by executing:

```
go get github.com/Shuttl-Tech/corn
```

## Usage

Corn can be configured using command line flags. At the moment only command line flags are available and support for environment variables or file based configuration is not on the roadmap.

The sub-command `manager` relies on two command line flags `--cmd` and `--sched` to configure periodic jobs.

These flags can be repeated as many times as desired to specify more than one task. Keep in mind that the order of flags matters in following ways:

 1. `--cmd` flag must be specified before `--sched`
 1. Every `--cmd` flag must be followed by a `--sched` flag

Consider, for example, this command that configures corn to execute two tasks:

```sh
corn manager --cmd "ls -lh /home" --sched "* * * * *" --cmd "date -u" --sched "*/2 * * * *"
```

The first task will execute `ls -lah /home` every minute, and the second one will execute `date -u` every two minutes.

## Contributing

Corn is licensed under the MIT License, and we greatly appreciate your contribution. You can contribute to corn by offering a pull request or opening an issue.

If you want to make code changes you will need Golang >= 1.14.

## License

Corn is released under the MIT License. A copy of the license is available in [LICENSE][] file.


[LICENSE]: ./LICENSE