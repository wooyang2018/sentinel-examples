# 快速开始

1、安装注册中心（可选）

安装consul：https://developer.hashicorp.com/consul/install

安装etcd：https://etcd.io/docs/v3.5/install/

2、启动注册中心

```bash
etcd 
consul agent -dev
```

etcd默认端口为2379，consul默认端口为8500

3、启动9个server进程

```
cd helloworld && ./setup.sh
```

4、客户端测试`helloworld/client/client_test.go`