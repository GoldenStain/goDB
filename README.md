# goDB

可以一个服务一个服务地实现前端接口，然后测试完再验证下一个。
可以参考Go的测试文件，全部在services目录下，文件名`*_test.go`

## 1. 使用方式

安装go1.23.0
配置`server/config.json`

在项目根目录运行
```go
go mod tidy
```
然后cd 到server目录，运行`go run .`

请不要修改proto_define目录下的library.proto文件

## 2. 监听端口

后端监听端口:`localhost:50051`

## 3. proto数据类型定义

详见proto_define/library.proto
绝大多数的数据类型都和C#前端定义的一致了，如果不一致请尽量修改前端（）
注意这里的int都是int32
123行之前都是具体的message定义，后面是每个service和相关的request,response
## 4. 接口规范

### 1. book_service部分

#### CreateBook

**描述**: 创建一本新书。

**请求**: `CreateBookRequest`

| 字段名         | 类型   | 说明             |
| -------------- | ------ | ---------------- |
| book_no        | string | 书籍编号         |
| title          | string | 书籍标题         |
| publisher_name | string | 出版社名称       |
| authors        | string | 作者             |
| stock_quantity | int32  | 库存数量         |
| price          | int32  | 价格             |
| keywords       | string | 关键词           |

**响应**: `CreateBookResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "book_no": "B001",
  "title": "Book Title 1",
  "publisher_name": "Test Publisher",
  "authors": "Test Author",
  "stock_quantity": 50,
  "price": 100,
  "keywords": "Test Keywords"
}
```

响应：
```json
{
  "success": true,
  "feedback": "Book created successfully"
}
```



