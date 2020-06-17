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

### vhost.conf

```
map $http_user_agent $is_bot {
    default 0;
    ~[a-z]bot[^a-z] 1;
    ~[sS]pider[^a-z] 1;
    'Yahoo! Slurp China' 1;
    'Mediapartners-Google' 1;
    'YisouSpider' 1;
    'Baiduspider' 1;
    'Googlebot' 1;
    'MSNBot' 1;
    'Baiduspider-image' 1;
    'YoudaoBot' 1;
    'Sogou web spider' 1;
    'Sogou inst spider' 1;
    'Sogou spider2' 1;
    'Sogou blog' 1;
    'Sogou News Spider' 1;
    'Sogou Orion spider' 1;
    'ChinasoSpider' 1;
    'Sosospider' 1;
    'EasouSpider' 1;
}
server {
    listen 80;
    server_name www.your-site-domain.com;
    root "/path-to-your-site";


    error_page 418 =200 @for_bots;
    if ($is_bot) {
        return 418;
    }

    location @for_bots {
        proxy_pass http://127.0.0.1:8899;
    }
}
```

## Keywords
**SEO** **golang** **go**