## Make your site friendly to search engine
Render website like a browser and return html without any external dependencies
- css rendering
- js rendering
## 让你的网站对搜索引擎友好
模拟浏览器渲染网页，让前后端分离的网站可以被搜索引擎检索，无需任何外部工具依赖
- css 渲染
- js 渲染

## Usages

### Run in code

```go
//go get -u github.com/jianzhiyao/sgo
rd := sgo.NewRender(sgo.Config{
		CacheSize: 1000,
        WaitTime:  time.Duration(WaitSecond),
        //for cache 3 seconds(0 is unlimited)
        CacheTime: 3,
})

response, hitCache, err := rd.GetSSR(backendUrl)
```

### Run  as a simple server
```command
cd cmd

//for windows
go build -o sgo-server.exe
sgo-server.exe -b http://127.0.0.1:8080 -p 8899 -w 3
//end for windows

//for linux
go build -o sgo-server
sgo-server -b http://127.0.0.1:8080 -p 8899 -w 3
//end for linux
```

## Keywords
**SEO** **golang** **go**
