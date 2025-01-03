# goDB

- 技术架构：前后端分离，并且通过gRPC交换protobuf字节流来实现高效率IO通信
- 实现语言：前端部分使用C#，借助Unity引擎实现；后端部分使用GOlang，基于GORM实现
- 技术特点：
    1. **lazy tag标记更新客户信用等级**

        timing计时器到时后，
        不立即对客户的相关信息进行更新，”only update on use”，只有在用到客户信息，例如调用GetCustomer方法时，才对信息进行更新

    2. **用虚拟客户充当计时器**

        不引入额外的即时逻辑，
        在数据库初始化时，创建一个虚拟客户，用其UpdateTime充当计时器，
        并在所有请求中手动排除，
        避免污染结果。

    3. **sending Email on update**

        检测所有缺书登记的
        Update、Delete操作，
        即时触发sending Email逻辑，第一时间通知客户。

    4. **多关键词检索、自定义模糊匹配**
    
        OnlineService模块，对用户输入的字符串进行多关键字匹配；在书籍检索中，能够根据自定义的匹配度阈值过滤搜索结果，可在json配置中更改阈值大小

## 前端部分

## 后端部分

- [goDB](#godb)
  - [前端部分](#前端部分)
  - [后端部分](#后端部分)
    - [1. 使用方式](#1-使用方式)
    - [2. 监听端口](#2-监听端口)
    - [3. proto数据类型定义](#3-proto数据类型定义)
    - [4. 接口规范](#4-接口规范)
      - [1. 书籍管理，第一个子页面，book\_service部分](#1-书籍管理第一个子页面book_service部分)
        - [CreateBook](#createbook)
        - [GetBook](#getbook)
        - [UpdateBook](#updatebook)
        - [DeleteBook](#deletebook)
      - [2. 缺书登记，第二个子页面，stock\_request\_service部分](#2-缺书登记第二个子页面stock_request_service部分)
        - [CreateStockRequest](#createstockrequest)
        - [UpdateStockRequest](#updatestockrequest)
        - [GetStockRequest](#getstockrequest)
        - [DeleteStockRequest](#deletestockrequest)
      - [3. 采购管理，第二个子页面，purchase\_order\_service部分](#3-采购管理第二个子页面purchase_order_service部分)
        - [CreatePurchaseOrder](#createpurchaseorder)
        - [GetPurchaseOrder](#getpurchaseorder)
        - [UpdatePurchaseOrder](#updatepurchaseorder)
        - [DeletePurchaseOrder](#deletepurchaseorder)
        - [GeneratePurchaseOrdersFromStockRequests](#generatepurchaseordersfromstockrequests)
      - [4. 客户管理，第三个子页面，customer\_service](#4-客户管理第三个子页面customer_service)
        - [CreateCustomer](#createcustomer)
        - [GetCustomer](#getcustomer)
        - [UpdateCustomer](#updatecustomer)
        - [DeleteCustomer](#deletecustomer)
      - [5. 订单、发货管理，第四个子页面，customer\_order\_service](#5-订单发货管理第四个子页面customer_order_service)
        - [CreateCustomerOrder](#createcustomerorder)
        - [GetCustomerOrder](#getcustomerorder)
        - [UpdateCustomerOrder](#updatecustomerorder)
        - [DeleteCustomerOrder](#deletecustomerorder)
      - [6. 供应商管理，第五个子页面，`supplier_service` 部分](#6-供应商管理第五个子页面supplier_service-部分)
        - [CreateSupplier](#createsupplier)
        - [GetSupplier](#getsupplier)
        - [UpdateSupplier](#updatesupplier)
        - [DeleteSupplier](#deletesupplier)
      - [6-2. 供书单管理，第五个子页面二级目录（供书名单，书目目录）`supply_book_service` 部分](#6-2-供书单管理第五个子页面二级目录供书名单书目目录supply_book_service-部分)
        - [CreateSupplyBook](#createsupplybook)
        - [GetSupplyBooksBySupplier](#getsupplybooksbysupplier)
        - [GetSupplyBookByID](#getsupplybookbyid)
        - [UpdateSupplyBook](#updatesupplybook)
        - [DeleteSupplyBook](#deletesupplybook)
      - [7. 浏览查询，第六个子页面，online\_service](#7-浏览查询第六个子页面online_service)
        - [QueryCustomer](#querycustomer)
        - [QueryBook](#querybook)

### 1. 使用方式

安装go1.23.0
配置`server/config.json`

在项目根目录运行
```go
go mod tidy
```

然后就可以用`go run .`启动服务器了。

生成proto和grpc代码的方式：
```
protoc -I=.\proto_define --go_out=. --go-grpc_out=. .\proto_define\library.proto
```

### 2. 监听端口

后端监听端口:`localhost:50051`

### 3. proto数据类型定义

详见proto_define/library.proto
绝大多数的数据类型都和C#前端定义的一致了，如果不一致请尽量修改前端（）
注意这里的int都是int32
123行之前都是具体的message定义，后面是每个service和相关的request,response
### 4. 接口规范

部分接口可能没有对应的按钮，可以不用

在一个请求当中，如果某个键不传入值，那么protobuf会把他设置为默认值，对于int是0，所以所有的`Update方法`都会跳过等于0的int键值，或者内容为空的string键值。因此，**在Update类型的方法中，如果不修改某字段，请不要指定该字段的值，以免覆盖**。

基本上所有的Update方法在修改的时候，都是用数据库主键ID（不是message里面的ID,比如客户有一个线上ID，OnlineID，与此不同）作为索引，这个索引在Get方法的时候可以得到。
大概流程就是：进入一个子页面，比如客户目录，通过Get方法得到一系列结果，每个结果都是一个完整的数据类型（比如一个Book），这里面带了ID，如果是能够Update的项目，在进入子页面进行修改的时候，一定要在请求里带上ID（**是数据库里的主键ID，不是其他名字的ID**），否则不能识别是哪个记录。

#### 1. 书籍管理，第一个子页面，book_service部分

为了减少不必要的数据库表数目，如果有多个author或者多个keyword，请把它们用逗号连接起来，作为一个整体传递给Request.  返回的response也会用逗号分割多个author或者keyword

##### CreateBook

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

##### GetBook

**描述**: 根据范围获取书籍。没有任何查询条件限制，专门服务于`库存目录`按钮，可以每次获取一个范围内的书籍，显示出来，用户向下翻页或者手动点击某个按钮的时候，再去获取下一批。

**请求**: `GetBookRequest`

| 字段名 | 类型  | 说明           |
| ------ | ----- | -------------- |
| start  | int32 | 起始位置       |
| stop   | int32 | 结束位置       |

**响应**: `GetBookResponse`

| 字段名   | 类型    | 说明       |
| -------- | ------- | ---------- |
| books    | Book[]  | 书籍列表   |
| success  | bool    | 是否成功   |
| feedback | string  | 反馈信息   |

**示例**:

请求:
```json
{
  "start": 0,
  "stop": 10
}
```

响应：
```json
{
  "books": [
    {
      "id": 1,
      "book_no": "B001",
      "title": "Book Title 1",
      "publisher_name": "Test Publisher",
      "price": 100,
      "stock_quantity": 50,
      "keywords": "Test Keywords",
      "authors": "Test Author",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "success": true,
  "feedback": "Books retrieved successfully"
}
```

##### UpdateBook

**描述**: 更新现有书籍。

**请求**: `UpdateBookRequest`

| 字段名         | 类型   | 说明             |
| -------------- | ------ | ---------------- |
| book_no        | string | 书籍编号         |
| title          | string | 书籍标题         |
| publisher_name | string | 出版社名称       |
| authors        | string | 作者             |
| stock_quantity | int32  | 库存数量         |
| price          | double | 价格             |
| keywords       | string | 关键词           |

**响应**: `UpdateBookResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "book_no": "B001",
  "title": "Updated Book Title",
  "publisher_name": "Updated Publisher",
  "authors": "Updated Author",
  "stock_quantity": 100,
  "price": 150.5,
  "keywords": "Updated Keywords"
}
```

响应：
```json
{
  "success": true,
  "feedback": "Book updated successfully"
}
```

##### DeleteBook

**描述**: 删除书籍。

**请求**: `DeleteBookRequest`

| 字段名 | 类型  | 说明     |
| ------ | ----- | -------- |
| book_id| int32 | 书籍ID   |

**响应**: `DeleteBookResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "book_id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Book deleted successfully"
}
```

#### 2. 缺书登记，第二个子页面，stock_request_service部分

##### CreateStockRequest

**描述**: 创建缺书登记。

**请求**: `CreateStockRequestRequest`

| 字段名      | 类型   | 说明       |
| ----------- | ------ | ---------- |
| book_no     | string | 书籍编号   |
| title       | string | 书籍标题   |
| publisher   | string | 出版社名称 |
| supplier    | string | 供应商     |
| author      | string | 作者       |
| quantity    | int32  | 数量       |
| request_date| string | 请求日期   |

**响应**: `CreateStockRequestResponse`

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
  "publisher": "Test Publisher",
  "supplier": "Test Supplier",
  "author": "Test Author",
  "quantity": 50,
  "request_date": "2023-01-01"
}
```

响应：
```json
{
  "success": true,
  "feedback": "Stock request created successfully"
}
```

##### UpdateStockRequest

**描述**: 更新缺书登记。  
当 `finished` 字段为 `true` 时，系统会检查是否存在相关的客户订单：
- 如果有客户订单，系统会生成假电子邮件通知客户，并返回反馈信息：  
  `"已暂存电子邮件通知客户"`。
- 如果没有客户订单，系统会返回反馈信息：  
  `"没有客户创建了书目xxx的相关订单，无须发送邮件"`。

这个方法比较特殊，因为写的时候前端还没有对应的按钮，所以不像其他update，只能用id来查询，这里提供了各种可选的字段，如果多个字段都不为空，就会取交集。

---

**请求**: `UpdateStockRequestRequest`

| 字段名       | 类型   | 说明               |
| ------------ | ------ | ------------------ |
| `book_no`    | string | 书籍编号           |
| `title`      | string | 书籍标题           |
| `publisher`  | string | 出版社名称         |
| `supplier`   | string | 供应商             |
| `author`     | string | 作者               |
| `quantity`   | int32  | 数量               |
| `request_date` | string | 请求日期（YYYY-MM-DD） |
| `id`         | int32  | 缺书登记 ID         |
| `finished`   | bool   | 是否完成           |

---

**响应**: `UpdateStockRequestResponse`

| 字段名      | 类型   | 说明               |
| ----------- | ------ | ------------------ |
| `success`   | bool   | 是否成功           |
| `feedback`  | string | 反馈信息           |

---

**示例**:

**请求**:
```json
{
  "book_no": "B001",
  "title": "Updated Book Title",
  "publisher": "Updated Publisher",
  "supplier": "Updated Supplier",
  "author": "Updated Author",
  "quantity": 100,
  "request_date": "2023-01-02",
  "id": 1,
  "finished": true
}
```

**响应**:
```json
{
  "success": true,
  "feedback": "已暂存电子邮件通知客户"
}
```

---

##### GetStockRequest

**描述**: 获取缺书登记信息。  

---

**请求**: `GetStockRequestRequest`

| 字段名  | 类型   | 说明       |
| ------- | ------ | ---------- |
| `start` | int32  | 起始位置   |
| `stop`  | int32  | 结束位置   |

---

**响应**: `GetStockRequestResponse`

| 字段名           | 类型              | 说明               |
| ----------------- | ----------------- | ------------------ |
| `success`        | bool              | 是否成功           |
| `feedback`       | string            | 反馈信息           |
| `stock_requests` | array of `StockRequest` | 缺书登记列表       |

---

##### DeleteStockRequest

**描述**: 删除缺书登记。  
- 当删除未完成的缺书登记时，系统会生成假电子邮件通知客户，并返回反馈信息：  
  `"已暂存电子邮件通知客户"`。
- 如果缺书登记已完成，则不会生成假电子邮件。

---

**请求**: `DeleteStockRequestRequest`

| 字段名            | 类型   | 说明         |
| ------------------ | ------ | ------------ |
| `stock_request_id` | int32  | 缺书登记 ID   |

---

**响应**: `DeleteStockRequestResponse`

| 字段名     | 类型   | 说明       |
| ---------- | ------ | ---------- |
| `success`  | bool   | 是否成功   |
| `feedback` | string | 反馈信息   | 




#### 3. 采购管理，第二个子页面，purchase_order_service部分

##### CreatePurchaseOrder

**描述**: 创建采购单。

**请求**: `CreatePurchaseOrderRequest`

| 字段名      | 类型   | 说明       |
| ----------- | ------ | ---------- |
| book_no     | string | 书籍编号   |
| title       | string | 书籍标题   |
| publisher   | string | 出版社名称 |
| supplier    | string | 供应商     |
| author      | string | 作者       |
| quantity    | int32  | 数量       |
| order_date  | string | 订单日期   |

**响应**: `CreatePurchaseOrderResponse`

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
  "publisher": "Test Publisher",
  "supplier": "Test Supplier",
  "author": "Test Author",
  "quantity": 50,
  "order_date": "2023-01-01"
}
```

响应：
```json
{
  "success": true,
  "feedback": "Purchase order created successfully"
}
```

##### GetPurchaseOrder

**描述**: 获取采购单。同样是专门服务于页面展示，每次获取一页。

**请求**: `GetPurchaseOrderRequest`

| 字段名 | 类型  | 说明           |
| ------ | ----- | -------------- |
| start  | int32 | 起始位置       |
| stop   | int32 | 结束位置       |

**响应**: `GetPurchaseOrderResponse`

| 字段名         | 类型             | 说明       |
| -------------- | ---------------- | ---------- |
| purchase_orders| PurchaseOrder[]  | 采购单列表 |
| success        | bool             | 是否成功   |
| feedback       | string           | 反馈信息   |

**示例**:

请求:
```json
{
  "start": 0,
  "stop": 10
}
```

响应：
```json
{
  "purchase_orders": [
    {
      "id": 1,
      "book_no": "B001",
      "title": "Book Title 1",
      "publisher": "Test Publisher",
      "supplier": "Test Supplier",
      "author": "Test Author",
      "quantity": 50,
      "order_date": "2023-01-01",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "finished": false
    }
  ],
  "success": true,
  "feedback": "Purchase orders retrieved successfully"
}
```

##### UpdatePurchaseOrder

**描述**: 更新采购单。

**请求**: `UpdatePurchaseOrderRequest`

| 字段名      | 类型   | 说明       |
| ----------- | ------ | ---------- |
| id          | int32  | 采购单ID   |
| book_no     | string | 书籍编号   |
| title       | string | 书籍标题   |
| publisher   | string | 出版社名称 |
| supplier    | string | 供应商     |
| author      | string | 作者       |
| quantity    | int32  | 数量       |
| order_date  | string | 订单日期   |
| finished    | bool   | 是否完成   |

**响应**: `UpdatePurchaseOrderResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "id": 1,
  "book_no": "B001",
  "title": "Updated Book Title",
  "publisher": "Updated Publisher",
  "supplier": "Updated Supplier",
  "author": "Updated Author",
  "quantity": 100,
  "order_date": "2023-01-02",
  "finished": true
}
```

响应：
```json
{
  "success": true,
  "feedback": "Purchase order updated successfully"
}
```

##### DeletePurchaseOrder

**描述**: 删除采购单。

**请求**: `DeletePurchaseOrderRequest`

| 字段名 | 类型  | 说明     |
| ------ | ----- | -------- |
| id     | int32 | 采购单ID |

**响应**: `DeletePurchaseOrderResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示示例**:

请求:
```json
{
  "id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Purchase order deleted successfully"
}
```

##### GeneratePurchaseOrdersFromStockRequests

**描述**: 根据未完成的缺书记录生成采购单。所有`Finished`字段为`false`的缺书记录，都会被设置为`true`，并生成相应的采购单。

**请求**: `GeneratePurchaseOrdersRequest`

**响应**: `GeneratePurchaseOrdersResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{}
```

响应：
```json
{
  "success": true,
  "feedback": "Purchase orders generated successfully"
}
```

#### 4. 客户管理，第三个子页面，customer_service

##### CreateCustomer

**描述**: 创建新客户。

**请求**: `CreateCustomerRequest`

| 字段名         | 类型   | 说明             |
| -------------- | ------ | ---------------- |
| online_id      | string | 在线ID           |
| password       | string | 密码             |
| name           | string | 姓名             |
| address        | string | 地址             |
| account_balance| int32  | 账户余额         |
| credit_level   | int32  | 信用等级         |

**响应**: `CreateCustomerResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "online_id": "user123",
  "password": "password123",
  "name": "John Doe",
  "address": "123 Main St",
  "account_balance": 1000,
  "credit_level": 5
}
```

响应：
```json
{
  "success": true,
  "feedback": "Customer created successfully"
}
```

##### GetCustomer

**描述**: 获取客户详细信息。同样每次提取一批，服务于客户目录的展示。

**请求**: `GetCustomerRequest`

| 字段名 | 类型  | 说明           |
| ------ | ----- | -------------- |
| start  | int32 | 起始位置       |
| stop   | int32 | 结束位置       |

**响应**: `GetCustomerResponse`

| 字段名   | 类型      | 说明       |
| -------- | --------- | ---------- |
| customers| Customer[]| 客户列表   |
| success  | bool      | 是否成功   |
| feedback | string    | 反馈信息   |

**示例**:

请求:
```json
{
  "start": 0,
  "stop": 10
}
```

响应：
```json
{
  "customers": [
    {
      "id": 1,
      "online_id": "user123",
      "password": "password123",
      "name": "John Doe",
      "address": "123 Main St",
      "account_balance": 1000,
      "credit_level": 5,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "success": true,
  "feedback": "Customers retrieved successfully"
}
```

##### UpdateCustomer

**描述**: 更新客户信息。

**请求**: `UpdateCustomerRequest`

| 字段名         | 类型   | 说明             |
| -------------- | ------ | ---------------- |
| id             | int32  | 客户ID           |
| online_id      | string | 在线ID           |
| password       | string | 密码             |
| name           | string | 姓名             |
| address        | string | 地址             |
| account_balance| int32  | 账户余额         |
| credit_level   | int32  | 信用等级         |

**响应**: `UpdateCustomerResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "id": 1,
  "online_id": "user123",
  "password": "newpassword123",
  "name": "John Doe",
  "address": "123 Main St",
  "account_balance": 2000,
  "credit_level": 6
}
```

响应：
```json
{
  "success": true,
  "feedback": "Customer updated successfully"
}
```

##### DeleteCustomer

**描述**: 删除客户记录。

**请求**: `DeleteCustomerRequest`

| 字段名 | 类型  | 说明     |
| ------ | ----- | -------- |
| id     | int32 | 客户ID   |

**响应**: `DeleteCustomerResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Customer deleted successfully"
}
```

#### 5. 订单、发货管理，第四个子页面，customer_order_service

##### CreateCustomerOrder

**描述**: 创建客户订单。

**请求**: `CreateCustomerOrderRequest`

| 字段名             | 类型   | 说明             |
| ------------------ | ------ | ---------------- |
| order_date         | string | 订单日期         |
| customer_online_id | string | 客户在线ID       |
| book_no            | string | 书籍编号         |
| book_count         | int32  | 书籍数量         |
| price              | int32  | 价格             |
| address            | string | 地址             |
| status             | string | 订单状态         |

**响应**: `CreateCustomerOrderResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "order_date": "2023-01-01",
  "customer_online_id": "user123",
  "book_no": "B001",
  "book_count": 2,
  "price": 200,
  "address": "123 Main St",
  "status": "Pending"
}
```

响应：
```json
{
  "success": true,
  "feedback": "Customer order created successfully"
}
```

##### GetCustomerOrder

**描述**: 获取客户订单。同样每次提取一批，服务于订单目录的展示。

**请求**: `GetCustomerOrderRequest`

| 字段名 | 类型  | 说明           |
| ------ | ----- | -------------- |
| start  | int32 | 起始位置       |
| stop   | int32 | 结束位置       |

**响应**: `GetCustomerOrderResponse`

| 字段名         | 类型              | 说明       |
| -------------- | ----------------- | ---------- |
| customer_orders| CustomerOrder[]   | 订单列表   |
| success        | bool              | 是否成功   |
| feedback       | string            | 反馈信息   |

**示例**:

请求:
```json
{
  "start": 0,
  "stop": 10
}
```

响应：
```json
{
  "customer_orders": [
    {
      "id": 1,
      "order_date": "2023-01-01",
      "customer_online_id": "user123",
      "book_no": "B001",
      "book_count": 2,
      "price": 200,
      "address": "123 Main St",
      "status": "Pending",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "success": true,
  "feedback": "Customer orders retrieved successfully"
}
```

##### UpdateCustomerOrder

**描述**: 更新客户订单。

**请求**: `UpdateCustomerOrderRequest`

| 字段名             | 类型   | 说明             |
| ------------------ | ------ | ---------------- |
| id                 | int32  | 订单ID           |
| order_date         | string | 订单日期         |
| customer_online_id | string | 客户在线ID       |
| book_no            | string | 书籍编号         |
| book_count         | int32  | 书籍数量         |
| price              | int32  | 价格             |
| address            | string | 地址             |
| status             | string | 订单状态         |

**响应**: `UpdateCustomerOrderResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "id": 1,
  "order_date": "2023-01-02",
  "customer_online_id": "user123",
  "book_no": "B001",
  "book_count": 3,
  "price": 300,
  "address": "123 Main St",
  "status": "Shipped"
}
```

响应：
```json
{
  "success": true,
  "feedback": "Customer order updated successfully"
}
```

##### DeleteCustomerOrder

**描述**: 删除客户订单。

**请求**: `DeleteCustomerOrderRequest`

| 字段名 | 类型  | 说明     |
| ------ | ----- | -------- |
| id     | int32 | 订单ID   |

**响应**: `DeleteCustomerOrderResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**:

请求:
```json
{
  "id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Customer order deleted successfully"
}
```

以下是将所有内容规范为 Markdown 格式后的结果：

---

#### 6. 供应商管理，第五个子页面，`supplier_service` 部分

为了方便在数据库中进行管理，把供书记录从`Supplier`中单独分离出来，用数据库表`SupplyBook`进行管理。

##### CreateSupplier

**描述**: 创建供应商。可以在创建的时候就指定一些供书单，也可以创建供应商之后，再通过`CreateSupplyBook`来添加。

**请求**: `CreateSupplierRequest`

| 字段名       | 类型         | 说明         |
| ------------ | ------------ | ------------ |
| name         | string       | 供应商名称   |
| basic_info   | string       | 基本信息     |
| supply_info  | string       | 供应信息     |
| supply_books | SupplyBook[] | 供应的书籍   |

**响应**: `CreateSupplierResponse`

| 字段名   | 类型   | 说明       |
| -------- | ------ | ---------- |
| success  | bool   | 是否成功   |
| feedback | string | 反馈信息   |

**示例**：

请求：
```json
{
  "name": "Supplier 1",
  "basic_info": "Basic Info",
  "supply_info": "Supply Info",
  "supply_books": [
    {
      "book_no": "B001",
      "title": "Book Title 1",
      "publisher_name": "Test Publisher",
      "price": 100,
      "quantity": 50
    }
  ]
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supplier created successfully"
}
```

---

##### GetSupplier

**描述**: 获取供应商信息（支持分页）。

**请求**: `GetSupplierRequest`

| 字段名 | 类型  | 说明     |
| ------ | ----- | -------- |
| start  | int32 | 起始位置 |
| stop   | int32 | 结束位置 |

**响应**: `GetSupplierResponse`

| 字段名    | 类型         | 说明         |
| --------- | ------------ | ------------ |
| success   | bool         | 是否成功     |
| feedback  | string       | 反馈信息     |
| suppliers | Supplier[]   | 供应商列表   |

**示例**：

请求：
```json
{
  "start": 0,
  "stop": 10
}
```

响应：
```json
{
  "success": true,
  "feedback": "Suppliers retrieved successfully",
  "suppliers": [
    {
      "id": 1,
      "name": "Supplier 1",
      "basic_info": "Basic Info",
      "supply_info": "Supply Info",
      "supply_books": [
        {
          "id": 1,
          "book_no": "B001",
          "title": "Book Title 1",
          "publisher_name": "Test Publisher",
          "price": 100,
          "quantity": 50,
          "supplier_id": 1,
          "created_at": "2023-01-01T00:00:00Z",
          "updated_at": "2023-01-01T00:00:00Z"
        }
      ],
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

---

##### UpdateSupplier

**描述**: 更新供应商信息。

**请求**: `UpdateSupplierRequest`

| 字段名       | 类型         | 说明         |
| ------------ | ------------ | ------------ |
| id           | int32        | 供应商ID     |
| name         | string       | 供应商名称   |
| basic_info   | string       | 基本信息     |
| supply_info  | string       | 供应信息     |
| supply_books | SupplyBook[] | 供应的书籍   |

**响应**: `UpdateSupplierResponse`

| 字段名   | 类型   | 说明     |
| -------- | ------ | -------- |
| success  | bool   | 是否成功 |
| feedback | string | 反馈信息 |

**示例**：

请求：
```json
{
  "id": 1,
  "name": "Updated Supplier",
  "basic_info": "Updated Basic Info",
  "supply_info": "Updated Supply Info",
  "supply_books": [
    {
      "book_no": "B001",
      "title": "Updated Book Title",
      "publisher_name": "Updated Publisher",
      "price": 150,
      "quantity": 100
    }
  ]
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supplier updated successfully"
}
```

---

##### DeleteSupplier

**描述**: 删除供应商。

**请求**: `DeleteSupplierRequest`

| 字段名 | 类型  | 说明     |
| ------ | ----- | -------- |
| id     | int32 | 供应商ID |

**响应**: `DeleteSupplierResponse`

| 字段名   | 类型   | 说明     |
| -------- | ------ | -------- |
| success  | bool   | 是否成功 |
| feedback | string | 反馈信息 |

**示例**：

请求：
```json
{
  "id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supplier deleted successfully"
}
```

---

#### 6-2. 供书单管理，第五个子页面二级目录（供书名单，书目目录）`supply_book_service` 部分

##### CreateSupplyBook

**描述**: 创建供书记录，一定要带上supplier_id，用来和数据库中的供应商绑定。

**请求**: `CreateSupplyBookRequest`

| 字段名          | 类型   | 说明         |
| --------------- | ------ | ------------ |
| book_no         | string | 书籍编号     |
| title           | string | 书籍标题     |
| publisher_name  | string | 出版社名称   |
| price           | int32  | 价格         |
| quantity        | int32  | 数量         |
| supplier_id     | int32  | 供应商ID     |

**响应**: `CreateSupplyBookResponse`

| 字段名   | 类型   | 说明     |
| -------- | ------ | -------- |
| success  | bool   | 是否成功 |
| feedback | string | 反馈信息 |

**示例**：

请求：
```json
{
  "book_no": "B001",
  "title": "Book Title 1",
  "publisher_name": "Test Publisher",
  "price": 100,
  "quantity": 50,
  "supplier_id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supply book created successfully"
}
```

---

##### GetSupplyBooksBySupplier

**描述**: 获取供应商的供书记录。配合页面展示使用，点击进入一个供应商页面的时候，会显示该供应商的供书记录，此时可以调用这个方法。

**请求**: `GetSupplyBooksBySupplierRequest`

| 字段名      | 类型  | 说明     |
| ----------- | ----- | -------- |
| supplier_id | int32 | 供应商ID |

**响应**: `GetSupplyBooksBySupplierResponse`

| 字段名       | 类型         | 说明         |
| ------------ | ------------ | ------------ |
| success      | bool         | 是否成功     |
| feedback     | string       | 反馈信息     |
| supply_books | SupplyBook[] | 供书记录列表 |

**示例**：

请求：
```json
{
  "supplier_id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supply books retrieved successfully",
  "supply_books": [
    {
      "id": 1,
      "book_no": "B001",
      "title": "Book Title 1",
      "publisher_name": "Test Publisher",
      "price": 100,
      "quantity": 50,
      "supplier_id": 1,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

---

##### GetSupplyBookByID

**描述**: 根据ID获取供书记录。

**请求**: `GetSupplyBookByIDRequest`

| 字段名 | 类型  | 说明       |
| ------ | ----- | ---------- |
| id     | int32 | 供书记录ID |

**响应**: `GetSupplyBookByIDResponse`

| 字段名       | 类型       | 说明         |
| ------------ | ---------- | ------------ |
| success      | bool       | 是否成功     |
| feedback     | string     | 反馈信息     |
| supply_book  | SupplyBook | 供书记录     |

**示例**：

请求：
```json
{
  "id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supply book retrieved successfully",
  "supply_book": {
    "id": 1,
    "book_no": "B001",
    "title": "Book Title 1",
    "publisher_name": "Test Publisher",
    "price": 100,
    "quantity": 50,
    "supplier_id": 1,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

---

##### UpdateSupplyBook

**描述**: 更新供书记录。

**请求**: `UpdateSupplyBookRequest`

| 字段名         | 类型   | 说明         |
| -------------- | ------ | ------------ |
| id             | int32  | 供书记录ID   |
| book_no        | string | 书籍编号     |
| title          | string | 书籍标题     |
| publisher_name | string | 出版社名称   |
| price          | int32  | 价格         |


| quantity       | int32  | 数量         |
| supplier_id    | int32  | 供应商ID     |

**响应**: `UpdateSupplyBookResponse`

| 字段名   | 类型   | 说明     |
| -------- | ------ | -------- |
| success  | bool   | 是否成功 |
| feedback | string | 反馈信息 |

**示例**：

请求：
```json
{
  "id": 1,
  "book_no": "B001",
  "title": "Updated Book Title",
  "publisher_name": "Updated Publisher",
  "price": 150,
  "quantity": 100,
  "supplier_id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supply book updated successfully"
}
```

---

##### DeleteSupplyBook

**描述**: 删除供书记录。

**请求**: `DeleteSupplyBookRequest`

| 字段名 | 类型  | 说明       |
| ------ | ----- | ---------- |
| id     | int32 | 供书记录ID |

**响应**: `DeleteSupplyBookResponse`

| 字段名   | 类型   | 说明     |
| -------- | ------ | -------- |
| success  | bool   | 是否成功 |
| feedback | string | 反馈信息 |

**示例**：

请求：
```json
{
  "id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "Supply book deleted successfully"
}
```


#### 7. 浏览查询，第六个子页面，online_service

##### QueryCustomer

**描述**: 查询客户信息。

**请求**: `QueryCustomerRequest`

| 字段名        | 类型   | 说明             |
| ------------- | ------ | ---------------- |
| online_id     | string | 在线ID           |
| name          | string | 客户姓名         |
| address       | string | 客户地址         |
| order_id      | int32  | 订单ID           |

这些字段当中，非空字段都会参与查询，（int非0就是非空）

**响应**: `QueryCustomerResponse`

| 字段名    | 类型        | 说明       |
| --------- | ----------- | ---------- |
| customers | Customer[]  | 客户列表   |
| success   | bool        | 是否成功   |
| feedback  | string      | 反馈信息   |

**示例**:

请求:
```json
{
  "online_id": "user123",
  "name": "张三",
  "address": "北京市朝阳区",
  "order_id": 1
}
```

响应：
```json
{
  "success": true,
  "feedback": "客户信息检索成功",
  "customers": [
    {
      "id": 1,
      "online_id": "user123",
      "password": "password123",
      "name": "张三",
      "address": "北京市朝阳区",
      "account_balance": 1000,
      "credit_level": 5,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "customer_orders": [
        {
          "id": 1,
          "order_date": "2023-01-01",
          "customer_online_id": "user123",
          "book_no": "B001",
          "book_count": 2,
          "price": 200,
          "address": "北京市朝阳区",
          "status": "Pending",
          "created_at": "2023-01-01T00:00:00Z",
          "updated_at": "2023-01-01T00:00:00Z"
        }
      ]
    }
  ]
}
```

##### QueryBook

**描述**: 查询书籍信息。

**请求**: `QueryBookRequest`

只要相似度超过匹配阈值，就会视为匹配
如果使用了多个非空项目进行匹配，只有每一项都超过匹配阈值才会视为匹配.
如果有多个关键词或者作者，只要其中一个匹配，就认为关键词/作者成功匹配了。

| 字段名         | 类型   | 说明             |
| -------------- | ------ | ---------------- |
| book_no        | string | 书籍编号         |
| title          | string | 书籍标题         |
| publisher_name | string | 出版社名称       |
| keywords       | string | 关键词           |
| authors        | string | 作者             |
| match_threshold | int32 | 匹配阈值         |

**响应**: `QueryBookResponse`

| 字段名   | 类型      | 说明       |
| -------- | --------- | ---------- |
| books    | Book[]    | 书籍列表   |
| success  | bool      | 是否成功   |
| feedback | string    | 反馈信息   |

**示例**:

请求:
```json
{
  "book_no": "B001",
  "title": "书籍标题1",
  "publisher_name": "出版社A",
  "keywords": "关键词1",
  "authors": "作者A",
  "match_threshold": 2
}
```

响应：
```json
{
  "success": true,
  "feedback": "书籍信息检索成功",
  "books": [
    {
      "id": 1,
      "book_no": "B001",
      "title": "书籍标题1",
      "publisher_name": "出版社A",
      "price": 100,
      "keywords": "关键词1",
      "authors": "作者A",
      "stock_quantity": 50,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```


