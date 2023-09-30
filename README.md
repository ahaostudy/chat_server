# chat_server
Chat Server用于启动本地的ChatGPT代理服务

### 简单启动
```shell
./chat
```

### 指定参数启动
通过 `-help` 查看可选参数
```shell
> ./chat -help

Usage of D:\code\Golang\chat\chat.exe:
  -host string
        host (default "0.0.0.0")
  -port int
        port (default 8080)
  -proxy string
        openai proxy (default "https://api.openai.com/")
```
示例
```shell
./chat ./chat -port=9090 -proxy='https://api.my-openai.com/'
```
