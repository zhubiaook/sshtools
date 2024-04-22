# sshtools

sshtools是一些基于SSH协议的工具，当前包含以下两个工具：

* rexec 批量连接多台主机，并发执行命令或脚本

* rcp 批量连接多台主机，并发下载/上传文件或目录

## 编译

```
$ https://github.com/zhubiaook/sshtools.git
$ cd sshtools
$ make
```

## rexec

> 批量连接多台主机，并发执行命令或脚本

#### 功能特性

* 可并发连接多台主机，极大缩短批量连接主机所消耗的时间。

* 当连接上多台主机后，即可执行命令/脚本，多台主机之间的命令/脚本是并发执行的，有效缩短批量主机执行命令/脚本所消耗的时间

* 执行脚本文件的流程是：先使用sftp协议上传本地脚本文件到远程主机`/tmp/scripts`目录下，再执行这些脚本文件。

* 批量连接主机即可通过命令行，也可使用配置文件。

#### 使用指南

```bash
rexec --help
Execute command or script concurrently on multiple SSH servers

Usage:
  rexec [flags]

Flags:
  -a, --addrs string      'host:port,host:port,...', The ssh server addresses, the falg is mutually exclusive with other flag '--config'
      --cmd string        A command passed to the ssh server for execution, the flag is mutually exclusive with other flag '--filename'
  -c, --config string     The ssh server configuration file, the flag is mutually exclusive with other flag '--addrs'
  -f, --filename string   A script file passed to the ssh server for execution, the flag is mutually exclusive with other flag '--cmd'
  -h, --help              help for rexec
  -p, --password string   The ssh server password
  -u, --username string   The ssh server username (default "root")
```

通过配置文件连接多台主机，查看这些主机的hostname

1. 创建配置文件`config.yaml` 
   
   ```yaml
   addrs:
     - addr: 10.20.141.19:22
       username: root
       password: 123
     - addr: 10.20.141.20:22
       username: root
       password: 123
     - addr: 10.20.141.21:22
       username: root
       password: 123
   ```

2. 远程主机上执行命令
   
   ```bash
   rexec -c configs/config.yaml --cmd 'hostname'
   ```

通过命令行参数连接多台主机，执行脚本文件

1. 遍行要执行的脚本文件`test.sh` 
   
   ```bash
   #!/bin/bash
   echo "hello world"
   ```

2. 远程主机上执行脚本文件

```bash
rexec -a '10.20.141.19:22:10.20.141.19:22' -u root -p '123' -f ./test.sh
```

## rcp

> 批量连接多台主机，并发下载/上传文件或目录

#### 功能特性

- 可并发连接多台主机，极大缩短批量连接主机所消耗的时间。

- 多台主机之间可并发上传/下载文件或目录

- 批量连接主机即可通过命令行，也可使用配置文件。

#### 使用指南

```bash
rcp --help
Copy files form/to multiple SSH server

Usage:
  rcp [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download files form multiple SSH server
  help        Help about any command
  upload      Upload files to multiple SSH server

Flags:
  -a, --addrs string        'host:port,host:port,...', The ssh server addresses
  -c, --config string       The ssh server configuration file
  -h, --help                help for rcp
  -l, --localpath string    Local file or directory
  -p, --password string     The ssh server password
  -r, --remotepath string   Remote file or directory
  -u, --username string     The ssh server username (default "root")

Use "rcp [command] --help" for more information about a command.
```

通过配置文件连接多台主机，上传/下载目录文件到远程主机上

1. 创建配置文件`config.yaml`
   
   ```yaml
   addrs:
   - addr: 10.20.141.19:22
     username: root
     password: 123
   - addr: 10.20.141.20:22
     username: root
     password: 123
   - addr: 10.20.141.21:22
     username: root
     password: 123
   ```

2. 上传目录文件
   
   ```bash
   rcp upload -c configs/config.yaml -l /local/some/path -r /remote/some/path
   ```

3. 下载目录文件
   
   ```bash
   rcp download -c configs/config.yaml -l /local/some/path -r /remote/some/path
   ```
   
   从多台远程主机上下载目录文件到本地，为了区分是哪台主机下载的文件，本地目录会自动加上该主机的addr作为路径的一部分，比如从`10.20.141.19:22` 上下载的文件存放到本地目录`/tmp/`下，最终的父路径为：
   
   ```bash
   /tmp/10.20.141.19:22
   ```
