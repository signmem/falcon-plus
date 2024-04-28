### 打包
```
make build
```

### 配置

执行文件的路径新建目录 `config`，新建两个配置文件 `test.yaml` 和 `prd.yaml`

### 测试环境启动

```
nohup ./collector &
```

### 生产环境启动

```
nohup ./collector -e prd &
```
