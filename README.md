# nw

`github.com/hkloudou/nw/v3` —— 一个极简的 Go HTTP 客户端脚手架，面向 API 调用与爬虫场景。

两个文件、零依赖、不到 350 行：

| 文件 | 内容 |
|---|---|
| `nw.go` | `Client`、`Option`、`Response`、`StatusError`、`JSON[T]` |
| `jar.go` | 可 JSON 序列化的 `Jar`（实现 `http.CookieJar`） |

## 设计原则

* **数据流就是 Go 的 `(value, error)`**。没有 `Result[T]`、`Then`、`Catch`。
* **一个 `Client` 管一切**：`*http.Client`、cookie、默认请求头都在里面，所有请求都从它发出。
* **一种扩展点**：`Option = func(*http.Request)`。客户端级（`Use`）和单次调用级共用同一类型。
* **Cookie 天然可持久化**：`Jar` 直接 `json.Marshal` / `Save` / `Load`，登录态在进程间复用。

## 快速开始

```go
c := nw.New() // 自带 Jar、60s 超时、浏览器 User-Agent
ctx := context.Background()

// 一行完成请求 + 解码
user, err := nw.JSON[User](c.Get(ctx, "https://api.example.com/me"))

// POST JSON：struct / map 自动 Marshal，string / []byte 原样发送
resp, err := c.PostJSON(ctx, url, map[string]any{"name": "x"})
resp, err := c.PostJSON(ctx, url, `{"name":"x"}`)

// 表单
resp, err := c.PostForm(ctx, url, url.Values{"user": {"a"}, "pass": {"b"}})

// 任意方法 / body
resp, err := c.Send(ctx, http.MethodPut, url, bytes.NewReader(data))

// 自己构造 *http.Request
resp, err := c.Do(req)
```

`Response` 内嵌 `*http.Response`（`StatusCode`、`Header`、`Cookies()`……），body 已完整读入：

```go
resp.Body     // []byte
resp.String() // string
resp.JSON(&v) // 解码
```

## Option

```go
// 客户端级：每个请求都生效
c.Use(nw.Header("Authorization", "Bearer "+token))

// 单次调用级：后执行，覆盖客户端级
c.Get(ctx, url, nw.Query("page", "2"), nw.Header("X-Trace", id))

// 自定义签名等逻辑就是一个闭包
sign := func(r *http.Request) { r.Header.Set("X-Sign", signOf(r)) }
c.Use(sign)
```

## 错误处理

* 网络 / 超时 / 构造失败：返回 `(nil, err)`。
* HTTP 状态码 `>= 400`：返回 `(resp, *StatusError)`，body 仍可读。

```go
resp, err := c.Get(ctx, url)
var se *nw.StatusError
if errors.As(err, &se) {
    log.Println(se.StatusCode, se.String()) // StatusError 内嵌 *Response
}
```

`nw.JSON[T]` 会原样透传错误，所以 `user, err := nw.JSON[User](c.Get(...))` 只需判一次 `err`。

## Cookie 复用

```go
c := nw.New()
_ = c.Jar.Load("cookies.json") // 首次运行文件不存在，忽略即可

if _, err := c.PostForm(ctx, loginURL, form); err != nil { ... }

_ = c.Jar.Save("cookies.json") // 下次进程直接是登录态
```

也可以自己存到任何地方：

```go
b, _ := json.Marshal(c.Jar)   // [{"Name":"sid","Value":"...","Domain":".example.com","Path":"/","Expires":"..."}]
_ = json.Unmarshal(b, c.Jar)   // 合并进现有 Jar

c.Jar.All()    // 全部 cookie（含 Domain/Path/Expires 等完整属性）
c.Jar.Clear()  // 清空
```

Jar 规则（Netscape 约定）：`Domain` 以 `.` 开头的对子域生效，否则仅对该主机生效；`Max-Age` 落盘时换算为绝对 `Expires`；过期 cookie 在读取时惰性清理；不查公共后缀列表，服务端可以给自身或父域设置 cookie。

## 其它

```go
c.Debug = true                            // 用标准 log 打印每次请求、状态码、耗时、响应体
c.Proxy("http://127.0.0.1:7890")          // 走代理；c.Proxy("") 恢复环境变量代理
c.HTTP.Timeout = 10 * time.Second         // *http.Client 直接暴露，Transport / TLS 等自行设置
```

## 从 v2 迁移

| v2 | v3 |
|---|---|
| `nw.Get[T](url, mw, nw.WithClient(c.GetHTTPClient()))` | `nw.JSON[T](c.Get(ctx, url))` |
| `Middlewaves[T].UseRequest(fn)` | `c.Use(fn)`（`fn` 就是 `func(*http.Request)`） |
| `Middlewaves[T].UseDecode(fn)` | `resp.JSON(&v)` 或自行解析 `resp.Body` |
| `Result[T]` / `Then` / `Catch` | `(T, error)` |
| `nw.WithCodes([]int{200, 201})` | 状态 `< 400` 不报错；其余用 `errors.As(err, &se)` 判断 |
| `cookiejar.Jar` + `DeepCopyFrom` | `nw.Jar` + `json.Marshal` / `Save` / `Load` |
| `Client.LocalStorage*` | 已移除，自行保存业务状态 |
| `Client.WiseProxy(u)` | `c.Proxy("http://...")` |
