# SecuArchive Web - 安全服务报告归档系统

## 项目简介

SecuArchive Web 是基于原版 SecuArchive 桌面应用重新开发的 Web 版本，采用 Go 语言作为后端，提供了安全服务报告的在线归档管理功能。

## 技术栈

- **后端**: Go + Gin Web 框架
- **数据库**: SQLite (gorm)
- **认证**: JWT
- **前端**: HTML + JavaScript + Bootstrap 5

## 功能特性

- [x] 用户登录认证
- [x] 项目管理（支持合同所属、合同编号）
- [x] 项目支持导入ZIP报告压缩包，自动智能分类
- [x] 报告智能分类（渗透测试、代码审计、基线核查、漏洞扫描等）
- [x] 自动识别复测报告
- [x] 报告文件管理（导入、分类、归档）
- [x] 项目维度的报告组织
- [x] 基础搜索和筛选
- [x] 报告下载
- [x] 模板管理（导入/下载测试报告模板）
- [x] AI 智能助手 (多模型支持：OpenAI、Claude、DeepSeek、智谱AI、百度AI、阿里云、月之暗面等)
- [x] 支持自定义模型和API端点
- [x] 数据备份/导出功能
- [x] 数据导入/恢复功能

## 快速开始

### 编译项目

```bash
cd /root/workspace/SecuArchiveWeb
export PATH=/usr/local/go/bin:$PATH
go build -o secuarchive-web ./cmd/server
```

### 运行服务

```bash
./secuarchive-web
```

服务将在 `http://localhost:8080` 启动。

### 访问系统

1. 打开浏览器访问 `http://localhost:8080/login`
2. 使用以下账号登录：
   - 用户名: `admin`
   - 密码: `admin123`

## 菜单结构

1. **仪表盘** - 系统概览
2. **项目管理** - 项目管理，支持导入ZIP压缩包
3. **报告管理** - 报告列表，自动智能分类
4. **模板管理** - 测试报告模板管理
5. **AI助手** - AI智能对话，配置模型
6. **备份管理** - 数据备份与恢复
7. **系统设置** - 用户密码修改

## 报告智能分类

系统根据文件名自动分类：

| 报告类型 | 关键词 | 子分类 |
|---------|--------|--------|
| 渗透测试 | 渗透、pentest、Web渗透 | Web渗透、APP渗透、内网渗透、红队评估 |
| 代码审计 | 代码、审计、源码 | Web代码审计、APP代码审计、SDK审计 |
| 基础环境测试 | 基线、配置、漏洞扫描 | 基线检查、配置核查、漏洞扫描 |
| 应急响应 | 应急、事件、响应 | 事件分析、溯源分析、处置报告 |
| 风险评估 | 风险 | 资产识别、威胁评估、脆弱性评估 |
| 合规审计 | 合规、等保、ISO | 等保测评、ISO27001 |

**复测标识**：文件名包含"复测"自动标记

## AI 模型支持

支持的模型厂商：
- OpenAI (GPT-4o, GPT-4, GPT-3.5-turbo)
- Anthropic (Claude)
- Google (Gemini)
- Azure OpenAI
- DeepSeek
- 智谱AI (GLM)
- 百度AI (ERNIE)
- 阿里云 (通义千问)
- 月之暗面 (Kimi)
- Ollama (本地部署)
- 自定义

## API 接口

### 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/auth/login | 用户登录 |
| POST | /api/auth/register | 用户注册 |
| GET | /api/auth/me | 获取当前用户信息 |

### 项目接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/projects | 获取项目列表 |
| GET | /api/projects/:id | 获取项目详情 |
| POST | /api/projects | 创建项目 |
| POST | /api/projects/import-zip | 创建项目并导入ZIP报告 |
| PUT | /api/projects/:id | 更新项目 |
| DELETE | /api/projects/:id | 删除项目 |
| GET | /api/projects/categories | 获取项目分类 |

### 报告接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/reports | 获取报告列表 |
| GET | /api/reports/:id | 获取报告详情 |
| POST | /api/reports | 导入报告 |
| PUT | /api/reports/:id | 更新报告 |
| DELETE | /api/reports/:id | 删除报告 |
| GET | /api/reports/:id/download | 下载报告 |
| GET | /api/reports/categories | 获取报告分类 |

### 模板接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/templates | 获取模板列表 |
| GET | /api/templates/:id | 获取模板详情 |
| POST | /api/templates | 导入模板 |
| GET | /api/templates/:id/download | 下载模板 |
| DELETE | /api/templates/:id | 删除模板 |
| GET | /api/templates/categories | 获取模板分类 |

### 备份接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/backups | 获取备份列表 |
| POST | /api/backups | 创建备份 |
| POST | /api/backups/import | 导入备份 |
| POST | /api/backups/:id/restore | 恢复备份 |
| DELETE | /api/backups/:id | 删除备份 |

### AI 接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/ai/config | 获取 AI 配置 |
| POST | /api/ai/config | 保存 AI 配置 |
| GET | /api/ai/providers | 获取 AI 模型厂商列表 |
| POST | /api/ai/chat | AI 对话 |
| POST | /api/ai/chat-with-context | AI 对话(带上下文) |
| POST | /api/ai/test | 测试 AI 连接 |

## 项目结构

```
SecuArchiveWeb/
├── cmd/
│   └── server/
│       └── main.go          # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── handlers/
│   │   ├── auth.go          # 认证处理器
│   │   ├── project.go       # 项目处理器
│   │   ├── report.go        # 报告处理器
│   │   ├── template.go      # 模板处理器
│   │   ├── backup.go        # 备份处理器
│   │   └── ai.go            # AI 处理器
│   ├── middleware/
│   │   └── auth.go          # JWT 认证中间件
│   ├── models/
│   │   └── models.go        # 数据模型
│   ├── services/
│   │   └── database.go      # 数据库服务
│   └── utils/
│       └── crypto.go        # 密码加密工具
├── web/
│   └── js/
│       └── app.js           # 前端 JavaScript
├── data/                    # 数据存储目录
│   ├── reports/             # 报告文件
│   ├── backups/             # 备份文件
│   ├── templates/           # 模板文件
│   └── secuarchive.db       # SQLite 数据库
├── go.mod
├── go.sum
└── README.md
```

## 配置

默认配置：

- 服务端口: `8080`
- 数据目录: `./data`
- JWT 过期时间: 7 天

可以通过环境变量 `DATA_DIR` 自定义数据目录。

## 注意事项

1. 首次运行时会自动创建默认用户 (admin/admin123)
2. 报告文件存储在 `data/reports/` 目录
3. 备份文件存储在 `data/backups/` 目录
4. 模板文件存储在 `data/templates/` 目录
5. 所有数据存储在 SQLite 数据库 `data/secuarchive.db`
